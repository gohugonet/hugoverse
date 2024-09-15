package valueobject

import (
	"github.com/bep/gitmap"
	"path/filepath"
	"strings"
	"time"
)

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

func NewGitInfo(info gitmap.GitInfo) GitInfo {
	return GitInfo{
		Hash:            info.Hash,
		AbbreviatedHash: info.AbbreviatedHash,
		Subject:         info.Subject,
		AuthorName:      info.AuthorName,
		AuthorEmail:     info.AuthorEmail,
		AuthorDate:      info.AuthorDate,
		CommitDate:      info.CommitDate,
	}
}

type GitMap struct {
	ContentDir string
	Repo       *gitmap.GitRepo
}

func (g *GitMap) GetInfo(filename string) GitInfo {
	name := strings.TrimPrefix(filepath.ToSlash(filename), g.ContentDir)
	name = strings.TrimPrefix(name, "/")
	gi, found := g.Repo.Files[name]
	if !found {
		return GitInfo{}
	}
	return NewGitInfo(*gi)
}
