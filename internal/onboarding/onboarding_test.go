package onboarding

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/ccw/ccw/internal/config"
)

func TestNeedsOnboarding(t *testing.T) {
	cfg := config.Default()
	if !NeedsOnboarding(cfg) {
		t.Fatal("default config should need onboarding")
	}

	cfg.Onboarded = true
	if NeedsOnboarding(cfg) {
		t.Fatal("onboarded config should not need onboarding")
	}
}

func TestAskStringDefault(t *testing.T) {
	store, err := config.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	stdin := strings.NewReader("\n")
	stdout := &bytes.Buffer{}
	o := NewWithIO(store, stdin, stdout)

	scanner := bufioScanner(stdin)
	result := o.askString(scanner, "Question?", "Help text", "default-value")
	if result != "default-value" {
		t.Fatalf("expected default-value, got %s", result)
	}
}

func TestAskStringCustom(t *testing.T) {
	store, err := config.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	stdin := strings.NewReader("custom-value\n")
	stdout := &bytes.Buffer{}
	o := NewWithIO(store, stdin, stdout)

	scanner := bufioScanner(stdin)
	result := o.askString(scanner, "Question?", "Help text", "default-value")
	if result != "custom-value" {
		t.Fatalf("expected custom-value, got %s", result)
	}
}

func TestAskBool(t *testing.T) {
	store, err := config.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input    string
		defVal   bool
		expected bool
	}{
		{"y\n", false, true},
		{"yes\n", false, true},
		{"n\n", true, false},
		{"no\n", true, false},
		{"\n", true, true},
		{"\n", false, false},
	}

	for _, tc := range tests {
		stdin := strings.NewReader(tc.input)
		stdout := &bytes.Buffer{}
		o := NewWithIO(store, stdin, stdout)
		scanner := bufioScanner(stdin)

		result := o.askBool(scanner, "Q", "H", tc.defVal)
		if result != tc.expected {
			t.Errorf("input=%q defVal=%v: expected %v, got %v",
				tc.input, tc.defVal, tc.expected, result)
		}
	}
}

func TestAskLayout(t *testing.T) {
	store, err := config.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Test default (option 1)
	stdin := strings.NewReader("\n")
	stdout := &bytes.Buffer{}
	o := NewWithIO(store, stdin, stdout)
	scanner := bufioScanner(stdin)

	layout := o.askLayout(scanner)
	if layout.Left != "claude" || layout.Right != "lazygit" {
		t.Fatalf("expected default layout, got %+v", layout)
	}

	// Test option 2
	stdin = strings.NewReader("2\n")
	o = NewWithIO(store, stdin, stdout)
	scanner = bufioScanner(stdin)

	layout = o.askLayout(scanner)
	if layout.Left != "lazygit" || layout.Right != "claude" {
		t.Fatalf("expected reversed layout, got %+v", layout)
	}
}

func bufioScanner(r *strings.Reader) *bufio.Scanner {
	return bufio.NewScanner(r)
}
