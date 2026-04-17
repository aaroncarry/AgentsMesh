package gitops

import (
	"strings"
	"testing"
)

func TestParseStatusOutput(t *testing.T) {
	output := strings.Join([]string{
		"M  backend/main.go",
		" M README.md",
		"R  old.txt -> new.txt",
		"?? untracked.txt",
		"D  deleted.txt",
	}, "\n")

	files, stats, hasStaged := parseStatusOutput(output)

	if len(files) != 5 {
		t.Fatalf("len(files) = %d, want 5", len(files))
	}
	if !hasStaged {
		t.Fatal("expected hasStaged to be true")
	}
	if stats.GetModified() != 2 {
		t.Fatalf("modified = %d, want 2", stats.GetModified())
	}
	if stats.GetRenamed() != 1 {
		t.Fatalf("renamed = %d, want 1", stats.GetRenamed())
	}
	if stats.GetUntracked() != 1 {
		t.Fatalf("untracked = %d, want 1", stats.GetUntracked())
	}
	if stats.GetDeleted() != 1 {
		t.Fatalf("deleted = %d, want 1", stats.GetDeleted())
	}
	if files[2].GetPath() != "new.txt" {
		t.Fatalf("renamed path = %q, want %q", files[2].GetPath(), "new.txt")
	}
}

func TestBuildPushEnvRejectsEmbeddedCredentials(t *testing.T) {
	_, err := buildPushEnv("https://user:pass@example.com/repo.git", "oauth2", "token")
	if err == nil {
		t.Fatal("expected error for embedded credentials")
	}
	if cmdErr, ok := err.(*CommandError); !ok || cmdErr.Code != "invalid_remote_url" {
		t.Fatalf("err = %#v, want invalid_remote_url", err)
	}
}
