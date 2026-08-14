package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallScriptVerifiesReleaseChecksum(t *testing.T) {
	const assetName = "reclaimit_v1.2.3_linux_amd64.tar.gz"
	const assetURL = "https://downloads.test/" + assetName
	const checksumURL = "https://downloads.test/reclaimit_v1.2.3_checksums.txt"

	tests := []struct {
		name        string
		checksumURL string
		checksum    func(string) string
		wantSuccess bool
		wantOutput  string
	}{
		{
			name:        "valid manifest",
			checksumURL: checksumURL,
			checksum: func(hash string) string {
				return hash + "  " + assetName + "\n"
			},
			wantSuccess: true,
			wantOutput:  "Verified SHA-256 checksum",
		},
		{
			name:        "missing manifest asset",
			checksumURL: "",
			checksum:    func(string) string { return "" },
			wantOutput:  "No checksum manifest found",
		},
		{
			name:        "malformed manifest",
			checksumURL: checksumURL,
			checksum:    func(string) string { return "not-a-checksum  " + assetName + "\n" },
			wantOutput:  "no valid SHA-256 entry",
		},
		{
			name:        "checksum mismatch",
			checksumURL: checksumURL,
			checksum:    func(string) string { return strings.Repeat("0", 64) + "  " + assetName + "\n" },
			wantOutput:  "Checksum mismatch",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			workspace := t.TempDir()
			archivePath := filepath.Join(workspace, assetName)
			writeInstallerArchive(t, archivePath)
			archive, err := os.ReadFile(archivePath)
			if err != nil {
				t.Fatal(err)
			}
			hash := fmt.Sprintf("%x", sha256.Sum256(archive))
			checksumPath := filepath.Join(workspace, "checksums.txt")
			if err := os.WriteFile(checksumPath, []byte(tc.checksum(hash)), 0o644); err != nil {
				t.Fatal(err)
			}

			fakeBin := filepath.Join(workspace, "bin")
			if err := os.Mkdir(fakeBin, 0o755); err != nil {
				t.Fatal(err)
			}
			writeExecutable(t, filepath.Join(fakeBin, "uname"), `#!/bin/sh
case "$1" in
  -s) printf 'Linux\n' ;;
  -m) printf 'x86_64\n' ;;
  *) exit 1 ;;
esac
`)
			writeExecutable(t, filepath.Join(fakeBin, "jq"), `#!/bin/sh
printf '%s\t%s\t%s\n' "$FAKE_ASSET_NAME" "$FAKE_ASSET_URL" "$FAKE_CHECKSUM_URL"
`)
			writeExecutable(t, filepath.Join(fakeBin, "curl"), `#!/bin/sh
output=''
url=''
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      output=$2
      shift 2
      ;;
    -*) shift ;;
    *)
      url=$1
      shift
      ;;
  esac
done
if [ -z "$output" ]; then
  printf '{}\n'
elif [ "$url" = "$FAKE_ASSET_URL" ]; then
  cp "$FAKE_ARCHIVE_PATH" "$output"
elif [ "$url" = "$FAKE_CHECKSUM_URL" ]; then
  cp "$FAKE_CHECKSUM_PATH" "$output"
else
  printf 'unexpected URL: %s\n' "$url" >&2
  exit 22
fi
`)

			installDir := filepath.Join(workspace, "install")
			cmd := exec.Command("bash", "./install.sh", "v1.2.3")
			cmd.Dir = repoRoot()
			cmd.Env = append(os.Environ(),
				"PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"),
				"HOME="+workspace,
				"RECLAIMIT_INSTALL_DIR="+installDir,
				"FAKE_ASSET_NAME="+assetName,
				"FAKE_ASSET_URL="+assetURL,
				"FAKE_CHECKSUM_URL="+tc.checksumURL,
				"FAKE_ARCHIVE_PATH="+archivePath,
				"FAKE_CHECKSUM_PATH="+checksumPath,
			)
			output, runErr := cmd.CombinedOutput()
			if tc.wantSuccess && runErr != nil {
				t.Fatalf("installer failed: %v\n%s", runErr, output)
			}
			if !tc.wantSuccess && runErr == nil {
				t.Fatalf("installer unexpectedly succeeded: %s", output)
			}
			if !strings.Contains(string(output), tc.wantOutput) {
				t.Fatalf("output missing %q: %s", tc.wantOutput, output)
			}

			target := filepath.Join(installDir, "reclaimit")
			if tc.wantSuccess {
				if info, err := os.Stat(target); err != nil || info.Mode().Perm() != 0o755 {
					t.Fatalf("installed binary is missing or has wrong mode: info=%v err=%v", info, err)
				}
			} else if _, err := os.Stat(installDir); !os.IsNotExist(err) {
				t.Fatalf("failed verification wrote to install directory: %v", err)
			}
		})
	}
}

func writeInstallerArchive(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gzipWriter := gzip.NewWriter(file)
	tarWriter := tar.NewWriter(gzipWriter)
	contents := []byte("#!/bin/sh\necho reclaimit test binary\n")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "reclaimit", Mode: 0o755, Size: int64(len(contents))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(contents); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeExecutable(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}
}
