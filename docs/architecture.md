# Architecture and Execution Flows

This document explains how `reclaimit` is structured and how the main commands move through the code.

## Scope

`reclaimit` is a single-binary Go CLI focused on developer-workstation cleanup analysis. The core behavior lives in a small set of modules:

- `cmd/reclaimit/main.go` is the executable entrypoint.
- `main.go` in the root package `reclaimit` orchestrates commands and output.
- scanner logic is split across `scanner_types.go`, `scanner_categories.go`, `scanner_analyze.go`, `scanner_summaries.go`, and `scanner_clean.go`.
- `selection.go` filters candidates after CLI or TUI selection.
- `tui.go` lets the user review and exclude targets interactively.
- `format.go` renders the final report.
- `internal/tui/tui.go` contains the reusable internal TUI package.
- `filesystem_unix.go` and `filesystem_windows.go` isolate per-platform filesystem usage data.

## C4: System Context

```mermaid
C4Context
    title reclaimit - System Context

    Person(developer, "Developer", "Runs cleanup analysis and decides what to delete")
    System(cli, "reclaimit CLI", "Analyzes reclaimable disk usage and optionally deletes reviewed targets")
    System_Ext(filesystem, "Local filesystem", "Developer home, repositories, caches, build artifacts")
    System_Ext(terminal, "Terminal / shell", "CLI execution and plain-text or Markdown output")
    System_Ext(tui, "Interactive TUI", "Tree-based review and exclusion UI")
    System_Ext(report, "Generated report", "Plain text or Markdown written to stdout or file")

    Rel(developer, terminal, "Runs commands in")
    Rel(developer, tui, "Reviews candidates in")
    Rel(terminal, cli, "Invokes")
    Rel(tui, cli, "Uses interactive command path")
    Rel(cli, filesystem, "Scans and optionally deletes known cleanup targets in")
    Rel(cli, report, "Produces")
    Rel(report, developer, "Documents reclaimable space for review")
```

## C4: Container View

```mermaid
C4Container
    title reclaimit - Containers

    Person(developer, "Developer")
    System_Boundary(reclaimit, "reclaimit") {
        Container(cli_router, "CLI router", "Go", "Parses flags, dispatches analyze / tui / clean")
        Container(scanner, "Scanner pipeline", "Go", "Walks directories, classifies candidates, aggregates summaries")
        Container(selection, "Selection pipeline", "Go", "Applies excluded groups and paths to the candidate set")
        Container(renderer, "Report renderer", "Go", "Renders plain text or Markdown output")
        Container(tui_container, "TUI", "tview / tcell", "Interactive tree for reviewing and excluding targets")
        Container(fs_adapter, "Filesystem usage adapter", "Go + OS APIs", "Collects total/free/available bytes per platform")
    }

    System_Ext(filesystem, "Local filesystem", "Directories, files, VCS markers, caches, artifacts")
    System_Ext(stdout, "Stdout / output file", "Console or saved report")

    Rel(developer, cli_router, "Runs")
    Rel(cli_router, scanner, "Calls Analyze")
    Rel(scanner, fs_adapter, "Reads disk usage")
    Rel(scanner, filesystem, "Traverses and inspects")
    Rel(cli_router, tui_container, "Opens for tui command")
    Rel(tui_container, selection, "Produces exclusions for")
    Rel(cli_router, selection, "Applies exclude flags or TUI selection through")
    Rel(cli_router, renderer, "Renders final report through")
    Rel(renderer, stdout, "Writes")
    Rel(cli_router, filesystem, "Deletes selected candidates during clean")
```

## C4: Component View

```mermaid
C4Component
    title reclaimit - Internal Components

    Container_Boundary(cli, "Go binary") {
        Component(main_entry, "Run / parseConfig", "main.go + options.go", "Routes commands and validates CLI configuration")
        Component(analyze_component, "Analyze", "scanner_analyze.go", "Builds Report from filesystem traversal")
        Component(scan_node, "AnalyzeWithOptions / scanContext.scan", "scanner_analyze.go", "Scanner core API and recursive traversal")
        Component(grouping, "determineGroup / findRepoRoot / ancestorGroup", "scanner_analyze.go", "Assigns candidate groups")
        Component(selection_component, "applySelection / filterCandidates", "selection.go", "Filters report by excluded groups and paths")
        Component(render_component, "RenderReport", "format.go", "Formats report as plain text or Markdown")
        Component(tui_component, "RunTUI / buildSelectionTree", "tui.go", "Interactive selection tree and selection snapshot")
        Component(fs_component, "filesystemUsage", "filesystem_*.go", "Returns filesystem capacity stats")
    }

    Rel(main_entry, analyze_component, "Calls")
    Rel(analyze_component, fs_component, "Uses")
    Rel(analyze_component, scan_node, "Delegates traversal to")
    Rel(scan_node, grouping, "Uses for each candidate")
    Rel(main_entry, tui_component, "Calls for tui")
    Rel(tui_component, selection_component, "Returns exclusions consumed by")
    Rel(main_entry, selection_component, "Uses through Analyze result")
    Rel(main_entry, render_component, "Uses for final output")
```

## Command Flow

```mermaid
flowchart TD
    A[User runs reclaimit] --> B{Command}
    B -->|analyze| C[parseConfig]
    B -->|tui| C
    B -->|clean| C
    C --> D[Analyze]
    D --> E[filesystemUsage]
    D --> F[scanContext.scan recursion]
    F --> G{Matched candidate?}
    G -->|yes| H[addCandidate]
    G -->|no| I[Continue traversal]
    H --> J[summarizeCategories / summarizeGroups]
    I --> J
    J --> K[applySelection]
    K --> L{Command is tui?}
    L -->|yes| M[RunTUI]
    M --> N[Collect excluded groups and paths]
    N --> O[RenderReport]
    L -->|no| P{Command is clean?}
    P -->|yes| Q[Preview deletion]
    Q --> R[Clean selected candidates]
    R --> S[Analyze again]
    S --> O
    P -->|no| O
    O --> T[stdout or --out file]
```

## Runtime Sequence

```mermaid
sequenceDiagram
    actor User
    participant CLI as Run(main.go)
    participant Analyze as Analyze
    participant Scan as scanContext.scan
    participant Select as applySelection
    participant TUI as RunTUI
    participant Render as RenderReport
    participant FS as Local filesystem

    User->>CLI: reclaimit analyze|tui|clean
    CLI->>Analyze: Analyze(cfg) -> AnalyzeWithOptions(...)
    Analyze->>FS: filesystemUsage(root)
    Analyze->>FS: ReadDir(root)
    loop per entry
        Analyze->>Scan: scan(path, inCandidateDir)
        Scan->>FS: Lstat / ReadDir
        Scan-->>Analyze: scanSummary + candidates
    end
    Analyze->>Select: applySelection(report, excludes)
    Select-->>Analyze: selected candidates and summaries
    alt tui command
        CLI->>TUI: RunTUI(report)
        TUI-->>CLI: selection snapshot
    end
    alt clean command
        CLI->>FS: RemoveAll(candidate.Path)
        CLI->>Analyze: Re-run after deletion
    end
    CLI->>Render: RenderReport(report, format)
    Render-->>User: plain text or Markdown report
```

## Candidate Detection and Safety Decisions

```mermaid
stateDiagram-v2
    [*] --> InspectNode
    InspectNode --> SkipSymlink: symlink
    InspectNode --> SkipSpecial: non-regular file
    InspectNode --> DirectoryNode: directory
    InspectNode --> FileNode: regular file

    DirectoryNode --> TraverseChildren
    TraverseChildren --> DirectoryCandidate: name matches category and size >= threshold
    TraverseChildren --> NoDirectoryCandidate: no match or below threshold
    DirectoryCandidate --> AddCandidate
    NoDirectoryCandidate --> Continue

    FileNode --> TrackTopFile
    TrackTopFile --> FileCandidate: extension matches and size >= threshold
    TrackTopFile --> NoFileCandidate: no match or below threshold
    FileCandidate --> AddCandidate
    NoFileCandidate --> Continue

    AddCandidate --> GroupCandidate
    GroupCandidate --> RepoGroup: group-mode=repo
    GroupCandidate --> DepthGroup: group-mode=depth
    RepoGroup --> SelectionFilter
    DepthGroup --> SelectionFilter
    SelectionFilter --> Excluded: excluded by group or exact path
    SelectionFilter --> Selected: kept for report or clean
    Excluded --> Continue
    Selected --> Continue
    SkipSymlink --> Continue
    SkipSpecial --> Continue
    Continue --> [*]
```

## Notes for Maintainers

- `Analyze` is the core module. Both `tui` and `clean` reuse it rather than implementing their own scan logic.
- `scanContext.scan` and helpers (`scanDir`, `scanFile`) concentrate most traversal behavior and remain the primary extension seam.
- Selection is a seam: exclusion logic is centralized in `selection.go` and reused by CLI flags and TUI output.
- The Windows filesystem adapter calls `GetDiskFreeSpaceExW` and returns real metrics with overflow-safe clamping.
