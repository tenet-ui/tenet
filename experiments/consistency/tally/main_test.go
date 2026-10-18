package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPack = "../../../packs/stripe-reference"

func TestSummarizeCountsPerCondition(t *testing.T) {
	rows := []row{
		{Condition: "plain", Brief: "a", Status: "ok", TokenOnly: false, Findings: 4},
		{Condition: "plain", Brief: "b", Status: "failed"},
		{Condition: "tools", Brief: "a", Status: "ok", TokenOnly: true, Findings: 0},
		{Condition: "tools", Brief: "b", Status: "ok", TokenOnly: true, Findings: 1},
	}
	want := []string{
		"- plain: 0 token-only of 2; 4 findings in total; 4.0 findings per ok output; 1 ok, 1 failed",
		"- tools: 2 token-only of 2; 1 findings in total; 0.5 findings per ok output; 2 ok, 0 failed",
	}
	got := summarize(rows)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("summarize =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestRenderShowsFailedRowWithDashes(t *testing.T) {
	out := render([]row{{Condition: "plain", Brief: "sign-in", Status: "failed"}})
	if !strings.Contains(out, "| plain | sign-in | failed | no | – | – |") {
		t.Fatalf("render output missing failed row:\n%s", out)
	}
}

func TestCollectRejectsStaleFailedFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "plain", "payments-list.html"), "<html></html>")
	writeFile(t, filepath.Join(dir, "plain", "payments-list.failed"), "boom")
	writeFile(t, filepath.Join(dir, "tools", "payments-list.html"), "<html></html>")
	_, err := collect(testPack, dir)
	if err == nil || !strings.Contains(err.Error(), "both .html and .failed") {
		t.Fatalf("collect error = %v, want a both-files error", err)
	}
}

func TestCollectMissingOutputsFails(t *testing.T) {
	_, err := collect(testPack, filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("collect with a missing outputs folder returned no error")
	}
}

func TestCollectCountsTokenOnlyAndFindings(t *testing.T) {
	dir := t.TempDir()
	clean := "<html><style>p { color: var(--color-ink); }</style></html>"
	dirty := "<html><style>p { color: #123456; }</style></html>"
	writeFile(t, filepath.Join(dir, "plain", "sign-in.html"), dirty)
	writeFile(t, filepath.Join(dir, "tools", "sign-in.html"), clean)
	rows, err := collect(testPack, dir)
	if err != nil {
		t.Fatalf("collect: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	plain, tools := rows[0], rows[1]
	if plain.TokenOnly || plain.Findings == 0 {
		t.Errorf("plain row = %+v, want findings and not token-only", plain)
	}
	if !tools.TokenOnly || tools.Findings != 0 || tools.Unresolved != 0 {
		t.Errorf("tools row = %+v, want token-only with no findings and no unresolved names", tools)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
