package scanner

import "time"

type Category struct {
	Key            string
	Display        string
	Description    string
	DirectoryNames map[string]struct{}
	FileExtensions map[string]struct{}
}

type PathSize struct {
	Path  string
	Bytes int64
}

type Candidate struct {
	Category    string
	CategoryKey string
	Path        string
	Group       string
	Bytes       int64
	Description string
	ModifiedAt  time.Time
	IsDir       bool
}

type CategorySummary struct {
	Category    string
	CategoryKey string
	Bytes       int64
	Count       int
	Description string
}

type GroupSummary struct {
	Group      string
	Bytes      int64
	Count      int
	ModifiedAt time.Time
}

type Report struct {
	Command                   string
	Root                      string
	TotalBytes                int64
	FreeBytes                 int64
	AvailableBytes            int64
	FilesystemBytes           int64
	TopEntries                []PathSize
	TopFiles                  []PathSize
	Candidates                []Candidate
	SelectedCandidates        []Candidate
	CandidateBytes            int64
	SelectedBytes             int64
	CategorySummaries         []CategorySummary
	GroupSummaries            []GroupSummary
	SelectedCategorySummaries []CategorySummary
	SelectedGroupSummaries    []GroupSummary
	DeletedBytes              int64
}
