package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunInitScaffoldsAValidFolder(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 0 {
		t.Fatalf("init exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "created design-system/") {
		t.Errorf("stdout = %q", stdout.String())
	}

	stdout.Reset()
	if code := run([]string{"validate", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 0 {
		t.Fatalf("validate exit = %d, stdout = %q", code, stdout.String())
	}
	want := "design system valid: 0 tokens, 0 rules, 0 components, 0 templates, 0 flows"
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
	}
}

func TestRunInitRefusesExistingFolder(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 0 {
		t.Fatalf("first init exit = %d", code)
	}
	stderr.Reset()
	if code := run([]string{"init", "--dir", root}, t.TempDir(), &stdout, &stderr); code != 1 {
		t.Fatalf("second init exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "already exists") {
		t.Errorf("stderr = %q, want it to say the folder already exists", stderr.String())
	}
}

func TestRunInitBadRoot(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"init", "--dir", "does-not-exist"}, t.TempDir(), &stdout, &stderr); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}
