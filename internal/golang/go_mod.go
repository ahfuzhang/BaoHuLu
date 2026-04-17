package golang

import (
	_ "embed"
	"fmt"
)

//go:embed go.mod.tpl
var goModTemplate string

// GoModContent renders the embedded go.mod template for a generated module.
func GoModContent(modulePath string) string {
	return fmt.Sprintf(goModTemplate, modulePath)
}
