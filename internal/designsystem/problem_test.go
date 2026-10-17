package designsystem

import (
	"testing"
)

func TestProblemString(t *testing.T) {
	if got := (Problem{File: "a.json", Line: 3, Message: "bad"}).String(); got != "a.json:3: bad" {
		t.Errorf("with line: got %q", got)
	}
	if got := (Problem{File: "a.json", Message: "missing"}).String(); got != "a.json: missing" {
		t.Errorf("without line: got %q", got)
	}
}

func TestLineAt(t *testing.T) {
	data := []byte("ab\ncd\nef")
	tests := []struct {
		off  int64
		want int
	}{
		{0, 1},
		{2, 1},
		{3, 2},
		{6, 3},
		{100, 3},
		{-1, 1},
	}
	for _, tt := range tests {
		if got := lineAt(data, tt.off); got != tt.want {
			t.Errorf("lineAt(%d) = %d, want %d", tt.off, got, tt.want)
		}
	}
}

func TestStripBOM(t *testing.T) {
	if got := string(stripBOM([]byte("\xEF\xBB\xBFx"))); got != "x" {
		t.Errorf("with BOM: got %q", got)
	}
	if got := string(stripBOM([]byte("x"))); got != "x" {
		t.Errorf("without BOM: got %q", got)
	}
}
