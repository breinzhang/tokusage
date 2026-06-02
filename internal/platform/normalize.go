package platform

import (
	"path"
	"strings"
)

func NormalizePathForStorage(input string) string {
	return path.Clean(input)
}

func NormalizeWindowsPathForStorage(input string) string {
	replaced := strings.ReplaceAll(input, `\`, `/`)
	if isWindowsDriveAbsolutePath(replaced) {
		drive := strings.ToLower(replaced[:1])
		cleanedPath := strings.ToLower(path.Clean(replaced[2:]))
		if cleanedPath == "/" {
			return drive + ":/"
		}
		return drive + ":" + cleanedPath
	}
	cleaned := path.Clean(replaced)
	return strings.ToLower(cleaned)
}

func isWindowsDriveAbsolutePath(path string) bool {
	if len(path) < 3 || path[1] != ':' || path[2] != '/' {
		return false
	}
	drive := path[0]
	return ('a' <= drive && drive <= 'z') || ('A' <= drive && drive <= 'Z')
}
