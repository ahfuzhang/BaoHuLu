package golang

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed go.mod.tpl
var goModTemplate string

// GoModContent renders the embedded go.mod template for a generated module.
// If withVtprotobuf is true, github.com/planetscale/vtprotobuf is added to the require block.
func GoModContent(modulePath string, withVtprotobuf bool) string {
	content := fmt.Sprintf(goModTemplate, modulePath)
	if withVtprotobuf {
		content = strings.Replace(content,
			"\tgithub.com/ahfuzhang/BaoHuLu v0.1.1",
			"\tgithub.com/ahfuzhang/BaoHuLu v0.1.1\n\tgithub.com/planetscale/vtprotobuf v0.6.0",
			1)
	}
	return content
}
