package designsystem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PointOfViewPath is the location of the point of view, relative to the project root.
const PointOfViewPath = "design-system/point-of-view.md"

var (
	// ErrPointOfViewMissing reports that the point of view file does not exist.
	ErrPointOfViewMissing = errors.New("point of view file not found")
	// ErrPointOfViewEmpty reports that the point of view file holds only whitespace.
	ErrPointOfViewEmpty = errors.New("point of view file is empty")
)

// ReadPointOfView returns the contents of root/design-system/point-of-view.md.
// The contents are returned unchanged, including any trailing newline.
func ReadPointOfView(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, PointOfViewPath))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("%w: %s", ErrPointOfViewMissing, PointOfViewPath)
		}
		return "", err
	}
	text := string(data)
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("%w: %s", ErrPointOfViewEmpty, PointOfViewPath)
	}
	return text, nil
}
