package claude

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestResolveProjectPrefersCWD(t *testing.T) {
	project := ResolveProject("/Users/example/work/repo", "/Users-brein-other")

	if project.Name != "repo" {
		t.Fatalf("Name = %q, want repo", project.Name)
	}
	if project.PathNorm != "/Users/example/work/repo" {
		t.Fatalf("PathNorm = %q", project.PathNorm)
	}
	if project.PathDisplay != "/Users/example/work/repo" {
		t.Fatalf("PathDisplay = %q", project.PathDisplay)
	}
	wantID := expectedProjectIDForTest("/Users/example/work/repo")
	if project.ID != wantID {
		t.Fatalf("ID = %q, want %q", project.ID, wantID)
	}
}

func TestResolveProjectFallsBackToTranscriptDir(t *testing.T) {
	project := ResolveProject("", "-Users-example-work-repo")

	if project.Name != "repo" {
		t.Fatalf("Name = %q, want repo", project.Name)
	}
	if project.PathNorm != "/Users/example/work/repo" {
		t.Fatalf("PathNorm = %q", project.PathNorm)
	}
	if project.PathDisplay != "/Users/example/work/repo" {
		t.Fatalf("PathDisplay = %q", project.PathDisplay)
	}
	wantID := expectedProjectIDForTest("/Users/example/work/repo")
	if project.ID != wantID {
		t.Fatalf("ID = %q, want %q", project.ID, wantID)
	}
}

func TestResolveProjectFallsBackToWindowsTranscriptDriveDir(t *testing.T) {
	project := ResolveProject("", "C--Users-example-work-repo")
	cwdProject := ResolveProject(`C:\Users\example\work\repo`, "")

	if project.Name != "repo" {
		t.Fatalf("Name = %q, want repo", project.Name)
	}
	if project.PathNorm != "c:/users/example/work/repo" {
		t.Fatalf("PathNorm = %q", project.PathNorm)
	}
	if project.PathDisplay != "C:/Users/example/work/repo" {
		t.Fatalf("PathDisplay = %q", project.PathDisplay)
	}
	if project.ID != cwdProject.ID {
		t.Fatalf("ID = %q, want Windows CWD ID %q", project.ID, cwdProject.ID)
	}
}

func TestResolveProjectNormalizesWindowsCWD(t *testing.T) {
	project := ResolveProject(`C:\Users\example\work\repo`, "")

	if project.Name != "repo" {
		t.Fatalf("Name = %q, want repo", project.Name)
	}
	if project.PathNorm != "c:/users/example/work/repo" {
		t.Fatalf("PathNorm = %q", project.PathNorm)
	}
	if project.PathDisplay != `C:\Users\example\work\repo` {
		t.Fatalf("PathDisplay = %q", project.PathDisplay)
	}
	wantID := expectedProjectIDForTest("c:/users/example/work/repo")
	if project.ID != wantID {
		t.Fatalf("ID = %q, want %q", project.ID, wantID)
	}
}

func TestResolveProjectWindowsEquivalentPathsShareStorageKey(t *testing.T) {
	projects := []struct {
		name        string
		cwd         string
		pathDisplay string
	}{
		{
			name:        "backslash uppercase drive",
			cwd:         `C:\Users\example\work\repo`,
			pathDisplay: `C:\Users\example\work\repo`,
		},
		{
			name:        "backslash lowercase path",
			cwd:         `c:\users\example\work\repo`,
			pathDisplay: `c:\users\example\work\repo`,
		},
		{
			name:        "slash uppercase drive",
			cwd:         `C:/Users/example/work/repo`,
			pathDisplay: `C:/Users/example/work/repo`,
		},
	}
	wantNorm := "c:/users/example/work/repo"
	wantID := expectedProjectIDForTest(wantNorm)

	for _, tt := range projects {
		t.Run(tt.name, func(t *testing.T) {
			project := ResolveProject(tt.cwd, "")

			if project.PathNorm != wantNorm {
				t.Fatalf("PathNorm = %q, want %q", project.PathNorm, wantNorm)
			}
			if project.ID != wantID {
				t.Fatalf("ID = %q, want %q", project.ID, wantID)
			}
			if project.PathDisplay != tt.pathDisplay {
				t.Fatalf("PathDisplay = %q, want %q", project.PathDisplay, tt.pathDisplay)
			}
		})
	}
}

func TestResolveProjectKeepsPosixBackslashLiteral(t *testing.T) {
	project := ResolveProject(`/tmp/foo\bar`, "")

	if project.Name != `foo\bar` {
		t.Fatalf("Name = %q, want %q", project.Name, `foo\bar`)
	}
	if project.PathNorm != `/tmp/foo\bar` {
		t.Fatalf("PathNorm = %q, want %q", project.PathNorm, `/tmp/foo\bar`)
	}
	if project.PathDisplay != `/tmp/foo\bar` {
		t.Fatalf("PathDisplay = %q, want %q", project.PathDisplay, `/tmp/foo\bar`)
	}
}

func TestResolveProjectIDStableAcrossSources(t *testing.T) {
	cwdProject := ResolveProject("/Users/example/work/repo", "")
	transcriptProject := ResolveProject("", "-Users-example-work-repo")

	if cwdProject.PathNorm != transcriptProject.PathNorm {
		t.Fatalf("PathNorm mismatch: cwd=%q transcript=%q", cwdProject.PathNorm, transcriptProject.PathNorm)
	}
	if cwdProject.ID != transcriptProject.ID {
		t.Fatalf("ID mismatch: cwd=%q transcript=%q", cwdProject.ID, transcriptProject.ID)
	}
}

func TestResolveProjectUnknown(t *testing.T) {
	project := ResolveProject("", "")

	if project.ID != "unknown-project" {
		t.Fatalf("ID = %q, want unknown-project", project.ID)
	}
}

func expectedProjectIDForTest(norm string) string {
	sum := sha256.Sum256([]byte("claude-code:" + norm))
	return hex.EncodeToString(sum[:])
}
