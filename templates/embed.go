package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

//go:embed all:golang all:csharp
var FS embed.FS

func Open(directory string) (fs.FS, error) {
	if strings.HasPrefix(directory, "builtin:") {
		name := strings.TrimPrefix(directory, "builtin:")
		if name == "." || !fs.ValidPath(name) {
			return nil, fmt.Errorf("invalid builtin bundle %q", name)
		}
		info, err := fs.Stat(FS, name)
		if err != nil {
			return nil, fmt.Errorf("unknown builtin bundle %q: %w", name, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("builtin bundle is not a directory: %s", name)
		}
		return fs.Sub(FS, name)
	}
	if directory == "" {
		return nil, fmt.Errorf("template directory is required")
	}
	info, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("template directory %s: %w", directory, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("template path is not a directory: %s", directory)
	}
	return os.DirFS(directory), nil
}
