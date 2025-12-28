package workspace

import (
	"fmt"
	"regexp"
	"strings"
)

var safeChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func WorkspaceID(repo, branch string) string {
	return fmt.Sprintf("%s/%s", repo, branch)
}

func SafeName(repo, branch string) string {
	id := WorkspaceID(repo, branch)
	name := strings.ReplaceAll(id, "/", "--")
	name = safeChars.ReplaceAllString(name, "-")
	if len(name) > 128 {
		name = name[:128]
	}
	name = strings.Trim(name, "-")
	if name == "" {
		return "workspace"
	}
	return name
}
