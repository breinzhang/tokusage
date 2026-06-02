package platform

import "testing"

func TestNormalizePathUsesSlashSeparators(t *testing.T) {
	got := NormalizePathForStorage(`/Users/example/work/repo`)
	want := `/Users/example/work/repo`
	if got != want {
		t.Fatalf("NormalizePathForStorage() = %q, want %q", got, want)
	}
}

func TestNormalizePathCleansPath(t *testing.T) {
	got := NormalizePathForStorage(`/Users/example/work/../work/repo/.`)
	want := `/Users/example/work/repo`
	if got != want {
		t.Fatalf("NormalizePathForStorage() = %q, want %q", got, want)
	}
}

func TestNormalizePathKeepsPosixBackslashLiteral(t *testing.T) {
	got := NormalizePathForStorage(`/tmp/foo\bar`)
	want := `/tmp/foo\bar`
	if got != want {
		t.Fatalf("NormalizePathForStorage() = %q, want %q", got, want)
	}
}

func TestNormalizeWindowsPathLowercasesPath(t *testing.T) {
	got := NormalizeWindowsPathForStorage(`C:\Users\example\work\repo`)
	want := `c:/users/example/work/repo`
	if got != want {
		t.Fatalf("NormalizeWindowsPathForStorage() = %q, want %q", got, want)
	}
}

func TestNormalizeWindowsPathCleansPath(t *testing.T) {
	got := NormalizeWindowsPathForStorage(`C:\Users\example\work\..\repo\.`)
	want := `c:/users/example/repo`
	if got != want {
		t.Fatalf("NormalizeWindowsPathForStorage() = %q, want %q", got, want)
	}
}

func TestNormalizeWindowsPathKeepsDriveRoot(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "backslash root",
			input: `C:\`,
			want:  `c:/`,
		},
		{
			name:  "slash root",
			input: `C:/`,
			want:  `c:/`,
		},
		{
			name:  "clamps above root",
			input: `C:\Users\..\..`,
			want:  `c:/`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeWindowsPathForStorage(tt.input)
			if got != tt.want {
				t.Fatalf("NormalizeWindowsPathForStorage() = %q, want %q", got, tt.want)
			}
		})
	}
}
