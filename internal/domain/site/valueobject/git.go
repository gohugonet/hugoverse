package valueobject

import "time"

// GitInfo provides information about a version controlled source file.
type GitInfo struct {
	// Commit hash.
	Hash string `json:"hash"`
	// Abbreviated commit hash.
	AbbreviatedHash string `json:"abbreviatedHash"`
	// The commit message's subject/title line.
	Subject string `json:"subject"`
	// The author name, respecting .mailmap.
	AuthorName string `json:"authorName"`
	// The author email address, respecting .mailmap.
	AuthorEmail string `json:"authorEmail"`
	// The author date.
	AuthorDate time.Time `json:"authorDate"`
	// The commit date.
	CommitDate time.Time `json:"commitDate"`
}
