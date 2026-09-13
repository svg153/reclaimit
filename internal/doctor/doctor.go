package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Check struct {
	Name    string
	Status  Status
	Message string
}

type Report struct {
	Version string
	GOOS    string
	GOARCH  string
	Checks  []Check
}

type Options struct {
	Version     string
	Args0       string
	Getwd       func() (string, error)
	UserHomeDir func() (string, error)
	LookPath    func(string) (string, error)
}

func Run(opts Options) Report {
	if opts.Getwd == nil {
		opts.Getwd = os.Getwd
	}
	if opts.UserHomeDir == nil {
		opts.UserHomeDir = os.UserHomeDir
	}
	if opts.LookPath == nil {
		opts.LookPath = exec.LookPath
	}

	report := Report{
		Version: opts.Version,
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
	}

	report.Checks = append(report.Checks, Check{
		Name:    "runtime",
		Status:  StatusOK,
		Message: runtime.GOOS + "/" + runtime.GOARCH,
	})

	if _, err := opts.Getwd(); err != nil {
		report.Checks = append(report.Checks, Check{
			Name:    "working-directory",
			Status:  StatusFail,
			Message: "current directory is not readable: " + err.Error(),
		})
	} else {
		report.Checks = append(report.Checks, Check{
			Name:    "working-directory",
			Status:  StatusOK,
			Message: "current directory is readable",
		})
	}

	if _, err := opts.UserHomeDir(); err != nil {
		report.Checks = append(report.Checks, Check{
			Name:    "home-directory",
			Status:  StatusWarn,
			Message: "home directory could not be detected: " + err.Error(),
		})
	} else {
		report.Checks = append(report.Checks, Check{
			Name:    "home-directory",
			Status:  StatusOK,
			Message: "home directory detected",
		})
	}

	binary := binaryName(opts.Args0)
	if _, err := opts.LookPath(binary); err != nil {
		report.Checks = append(report.Checks, Check{
			Name:    "path",
			Status:  StatusWarn,
			Message: fmt.Sprintf("%s is not reachable from PATH", binary),
		})
	} else {
		report.Checks = append(report.Checks, Check{
			Name:    "path",
			Status:  StatusOK,
			Message: fmt.Sprintf("%s is reachable from PATH", binary),
		})
	}

	return report
}

func Render(report Report) string {
	var b strings.Builder
	b.WriteString("reclaimit doctor\n")
	fmt.Fprintf(&b, "version: %s\n", valueOrUnknown(report.Version))
	fmt.Fprintf(&b, "runtime: %s/%s\n\n", report.GOOS, report.GOARCH)
	for _, check := range report.Checks {
		fmt.Fprintf(&b, "[%s] %s: %s\n", check.Status, check.Name, check.Message)
	}
	return b.String()
}

func ExitCode(report Report) int {
	for _, check := range report.Checks {
		if check.Status == StatusFail {
			return 1
		}
	}
	return 0
}

func binaryName(args0 string) string {
	args0 = strings.TrimSpace(args0)
	if args0 == "" {
		return "reclaimit"
	}
	parts := strings.FieldsFunc(args0, func(r rune) bool {
		return r == '/' || r == '\\'
	})
	if len(parts) == 0 {
		return "reclaimit"
	}
	name := parts[len(parts)-1]
	if name == "" {
		return "reclaimit"
	}
	return name
}

func valueOrUnknown(value string) string {
	if strings.TrimSpace(value) == "" {
		return "unknown"
	}
	return value
}
