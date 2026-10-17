package designsystem

import (
	"bytes"
	"fmt"
)

// Problem is one validation finding. Line is 1-based; 0 means the whole file.
type Problem struct {
	File    string
	Line    int
	Message string
}

// String formats the problem as "file:line: message", or "file: message" when Line is 0.
func (p Problem) String() string {
	if p.Line == 0 {
		return fmt.Sprintf("%s: %s", p.File, p.Message)
	}
	return fmt.Sprintf("%s:%d: %s", p.File, p.Line, p.Message)
}

// lineAt returns the 1-based line of byte offset off in data. Offsets outside data are clamped.
func lineAt(data []byte, off int64) int {
	if off < 0 {
		off = 0
	}
	if off > int64(len(data)) {
		off = int64(len(data))
	}
	return 1 + bytes.Count(data[:off], []byte("\n"))
}

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// stripBOM removes a leading UTF-8 byte-order mark. Line numbers are unaffected.
func stripBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, utf8BOM)
}
