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
# process signal/bootstrap boundary and test-support code. The testable command
# and cancellation behavior lives in the root package.
module="$(go list -m)"
readonly module
readonly -a excluded_packages=(
  "${module}/cmd/reclaimit"
  "${module}/internal/test"
  "${module}/internal/testhelpers"
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

# Enforce the same floor for every production source file. This catches a
# lightly tested high-risk file hidden by strong coverage elsewhere in its
# package. Bootstrap and test-support paths use the same exclusions as above.
if ! awk -v floor="${minimum}" -v module="${module}" '
  NR == 1 { next }
  {
    split($1, location, ":")
    file = location[1]
    total[file] += $2
    if ($3 > 0) covered[file] += $2
  }
  END {
    failed = 0
    for (file in total) {
      if (file == module "/cmd/reclaimit/main.go" ||
          index(file, module "/internal/test/") == 1 ||
          index(file, module "/internal/testhelpers/") == 1) {
        continue
      }
      coverage = 100 * covered[file] / total[file]
      display = file
      sub("^" module "/", "", display)
      printf "Coverage file: %s %.1f%%\n", display, coverage
      if (coverage + 0.000001 < floor) {
        printf "File coverage %.1f%% is below %.1f%%: %s\n", coverage, floor, display > "/dev/stderr"
        failed = 1
      }
    }
    exit failed
  }
' coverage.out; then
	failed=1
fi

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### Coverage"
    echo "* Total coverage: ${total}%"
    echo "* Minimum overall and per production package and file: ${minimum}%"
  } >> "${GITHUB_STEP_SUMMARY}"
fi

exit "${failed}"
