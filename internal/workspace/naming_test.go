package workspace

import "testing"

func TestSafeName(t *testing.T) {
	tests := []struct {
		name   string
		repo   string
		branch string
		expect string
	}{
		{"simple", "myrepo", "main", "myrepo--main"},
		{"with slash", "myrepo", "feature/test", "myrepo--feature--test"},
		{"repo with dot", "elicited.blog", "rss-feed", "elicited_blog--rss-feed"},
		{"multiple dots", "foo.bar.baz", "feature", "foo_bar_baz--feature"},
		{"dot in branch", "repo", "fix.bug", "repo--fix_bug"},
		{"special chars", "my@repo", "feat#1", "my-repo--feat-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeName(tt.repo, tt.branch)
			if got != tt.expect {
				t.Errorf("SafeName(%q, %q) = %q, want %q", tt.repo, tt.branch, got, tt.expect)
			}
		})
	}
}
