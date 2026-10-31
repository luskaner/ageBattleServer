package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// BuildCommit and BuildDirty are stamped in at link time by scripts/deploy.ps1:
//
//	go build -ldflags "-X main.BuildCommit=<sha> -X main.BuildDirty=<dirty>"
//
// so the running binary can say exactly which source it came from. A capture
// that does not record this is not reproducible, and the commit is the first
// thing anyone asks for when a capture turns out to disagree with another one.
var (
	BuildCommit = "dev"
	BuildDirty  = "unknown"
)

// CommitInfo is the provenance block written next to every capture.
type CommitInfo struct {
	Commit       string `json:"commit"`
	Short        string `json:"short"`
	Dirty        string `json:"dirty"`
	GoVersion    string `json:"go_version"`
	ProxySHA256  string `json:"proxy_sha256"`
	ServerSHA    string `json:"primary_server_sha256"`
	SecondarySHA string `json:"secondary_server_sha256"`
	When         string `json:"when"`
	Note         string `json:"note,omitempty"`
}

func shortCommit(c string) string {
	c = strings.TrimSpace(c)
	if len(c) > 12 {
		return c[:12]
	}
	return c
}

// WriteCommitFile writes COMMIT.txt into the capture directory. It is a plain
// text file on purpose: it has to be readable with `cat` on whatever machine the
// capture is copied to, years later, without this tool.
func WriteCommitFile(dir string, primary, secondary ExeInfo) error {
	info := CommitInfo{
		Commit:       BuildCommit,
		Short:        shortCommit(BuildCommit),
		Dirty:        BuildDirty,
		ServerSHA:    primary.SHA256,
		SecondarySHA: secondary.SHA256,
		When:         nowStamp(),
	}
	if BuildCommit == "dev" {
		info.Note = "built without -ldflags, so the commit is unknown; run scripts/deploy.ps1"
	}
	body := strings.Join([]string{
		"# provenance of this capture",
		"commit:      " + info.Commit,
		"dirty:       " + info.Dirty,
		"when:        " + info.When,
		"",
		"# identity is decided by CONTENT, not by these file names.",
		"# See roles.json in this directory, or README for the method.",
		"",
		"main_log      (T1) " + string(primary.Identity) + "  " + primary.SHA256 + "  " + primary.Path,
		"secondary_log (T2) " + string(secondary.Identity) + "  " + secondary.SHA256 + "  " + secondary.Path,
		"",
	}, "\r\n")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "COMMIT.txt"), []byte(body), 0o644)
}

// StampTo is a one line summary for the log, so a capture can be identified from
// proxy.log alone.
func StampTo() string {
	return fmt.Sprintf("commit=%s dirty=%s", shortCommit(BuildCommit), BuildDirty)
}
