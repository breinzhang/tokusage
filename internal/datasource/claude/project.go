package claude

import (
	"crypto/sha256"
	"encoding/hex"
	"path"
	"strings"

	"github.com/breinzhang/tokusage/internal/domain"
	"github.com/breinzhang/tokusage/internal/platform"
)

func ResolveProject(cwd string, transcriptDirName string) domain.Project {
	projectPath := cwd
	if projectPath == "" {
		projectPath = decodeTranscriptProjectDir(transcriptDirName)
	}
	if projectPath == "" {
		return domain.Project{
			ID:          "unknown-project",
			Name:        "unknown-project",
			PathNorm:    "",
			PathDisplay: "",
		}
	}

	norm := normalizeProjectPathForStorage(projectPath)
	name := path.Base(norm)
	if name == "." || name == "/" || name == "" {
		name = "unknown-project"
	}
	return domain.Project{
		ID:          projectID(norm),
		Name:        name,
		PathNorm:    norm,
		PathDisplay: projectPath,
	}
}

func normalizeProjectPathForStorage(projectPath string) string {
	if isWindowsProjectPath(projectPath) {
		return platform.NormalizeWindowsPathForStorage(projectPath)
	}
	return platform.NormalizePathForStorage(projectPath)
}

func isWindowsProjectPath(projectPath string) bool {
	return isWindowsDriveAbsolutePath(projectPath)
}

func isWindowsDriveAbsolutePath(projectPath string) bool {
	if len(projectPath) < 3 || projectPath[1] != ':' {
		return false
	}
	if projectPath[2] != '\\' && projectPath[2] != '/' {
		return false
	}
	drive := projectPath[0]
	return ('a' <= drive && drive <= 'z') || ('A' <= drive && drive <= 'Z')
}

func decodeTranscriptProjectDir(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if isWindowsTranscriptDriveDir(name) {
		return name[:1] + ":/" + strings.ReplaceAll(name[3:], "-", "/")
	}
	if strings.HasPrefix(name, "-") {
		return "/" + strings.ReplaceAll(strings.TrimPrefix(name, "-"), "-", "/")
	}
	return strings.ReplaceAll(name, "-", "/")
}

func isWindowsTranscriptDriveDir(name string) bool {
	if len(name) < 3 || name[1] != '-' || name[2] != '-' {
		return false
	}
	drive := name[0]
	return ('a' <= drive && drive <= 'z') || ('A' <= drive && drive <= 'Z')
}

func projectID(norm string) string {
	sum := sha256.Sum256([]byte("claude-code:" + norm))
	return hex.EncodeToString(sum[:])
}
