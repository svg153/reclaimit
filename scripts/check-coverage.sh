#!/usr/bin/env bash

set -euo pipefail

readonly minimum="${COVERAGE_MINIMUM:-90}"
repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly repository_root
temporary_directory="$(mktemp -d)"
readonly temporary_directory
trap 'rm -rf -- "${temporary_directory}"' EXIT

cd "${repository_root}"

go test ./... -covermode=atomic -coverprofile=coverage.out
go tool cover -func=coverage.out | tee coverage.txt

total="$({ awk '/^total:/ {gsub(/%/, "", $3); print $3}' coverage.txt; } | tail -n 1)"
if [[ -z "${total}" ]]; then
  echo "Unable to determine total coverage" >&2
  exit 1
fi

below_minimum() {
  awk -v coverage="$1" -v floor="${minimum}" 'BEGIN { exit coverage < floor ? 0 : 1 }'
}

failed=0
if below_minimum "${total}"; then
  echo "Total coverage ${total}% is below ${minimum}%" >&2
  failed=1
else
  echo "Total coverage: ${total}% (minimum ${minimum}%)"
fi

# These exclusions are not meaningful per-package unit-test targets: the
# process signal/bootstrap boundary, test-support code, and declarations-only
# types. The testable command and cancellation behavior lives in the root package.
module="$(go list -m)"
readonly module
readonly -a excluded_packages=(
  "${module}/cmd/reclaimit"
  "${module}/internal/test"
  "${module}/internal/testhelpers"
  "${module}/internal/types"
)

is_excluded() {
  local package="$1"
  local excluded
  for excluded in "${excluded_packages[@]}"; do
    if [[ "${package}" == "${excluded}" ]]; then
      return 0
    fi
  done
  return 1
}

while IFS= read -r package; do
  display_package="${package#"${module}"}"
  display_package="${display_package#/}"
  display_package="${display_package:-.}"
  if is_excluded "${package}"; then
    echo "Coverage package: ${display_package} (not applicable)"
    continue
  fi

  profile="${temporary_directory}/package.out"
  go test "${package}" -covermode=atomic -coverprofile="${profile}" >/dev/null
  package_coverage="$({ go tool cover -func="${profile}" | awk '/^total:/ {gsub(/%/, "", $3); print $3}'; } | tail -n 1)"

  if [[ -z "${package_coverage}" ]]; then
    echo "Unable to determine coverage for ${package}" >&2
    failed=1
  elif below_minimum "${package_coverage}"; then
    echo "Coverage package: ${display_package} ${package_coverage}% (below ${minimum}%)" >&2
    failed=1
  else
    echo "Coverage package: ${display_package} ${package_coverage}%"
  fi
done < <(go list ./...)

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### Coverage"
    echo "* Total coverage: ${total}%"
    echo "* Minimum overall and per production package: ${minimum}%"
  } >> "${GITHUB_STEP_SUMMARY}"
fi

exit "${failed}"
