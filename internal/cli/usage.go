package cli

import "strings"

func UsageText(topic string) string {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return `reclaimit scans a workspace for cleanup candidates.

Usage:
  reclaimit [command] [flags]

Commands:
  analyze    Scan and print a report
  clean      Delete the selected candidates after review
  tui        Open the interactive review interface
  help       Show command help
  version    Print the version

Use "reclaimit <command> --help" for more information about a command.
`
	}

	return `Usage:
  reclaimit ` + topic + ` [flags]

Flags:
  -root PATH
      path to scan
  -format plain|markdown
      output format
  -group-mode repo|depth
      group candidates by repository or path depth
  -group-depth N
      depth to use when group-mode=depth
  -top-files N
      number of largest files to show
  -top-groups N
      number of candidate groups to show
  -top-entries N
      number of largest direct children under root to show
  -min-candidate-size BYTES
      minimum candidate size in bytes
  -out FILE
      write the report to a file
  -include-category VALUE
      limit to a category (repeatable)
  -exclude-category VALUE
      exclude a category (repeatable)
  -exclude-group PATH
      exclude a group path prefix (repeatable)
  -exclude-path PATH
      exclude a specific candidate path (repeatable)
  -yes
      confirm destructive cleanup when using clean
  -log-level debug|info|warn|error
      log verbosity sent to stderr
`
}

func validLogLevel(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}
