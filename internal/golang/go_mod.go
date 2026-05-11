package golang

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed go.mod.tpl
var goModTemplate string

// sonicIndirectDeps lists the transitive (indirect) dependencies of github.com/bytedance/sonic v1.15.1.
// Pinned versions match the go.sum produced by `go mod tidy` with that release.
const sonicIndirectDeps = "\nrequire (\n" +
	"\tgithub.com/bytedance/gopkg v0.1.3 // indirect\n" +
	"\tgithub.com/bytedance/sonic/loader v0.5.1 // indirect\n" +
	"\tgithub.com/cloudwego/base64x v0.1.6 // indirect\n" +
	"\tgithub.com/klauspost/cpuid/v2 v2.2.9 // indirect\n" +
	"\tgithub.com/twitchyliquid64/golang-asm v0.15.1 // indirect\n" +
	"\tgolang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect\n" +
	"\tgolang.org/x/sys v0.22.0 // indirect\n" +
	")\n"

// GoModContent renders the embedded go.mod template for a generated module.
// If withVtprotobuf is true, github.com/planetscale/vtprotobuf is added to the require block.
// If withTest is true, github.com/bytedance/sonic and its indirect deps are added so the
// generated directory works without running go mod tidy.
func GoModContent(modulePath string, withVtprotobuf, withTest bool) string {
	content := fmt.Sprintf(goModTemplate, modulePath)
	if withVtprotobuf {
		content = strings.Replace(content,
			"\tgithub.com/ahfuzhang/BaoHuLu v0.4.3",
			"\tgithub.com/ahfuzhang/BaoHuLu v0.4.3\n\tgithub.com/planetscale/vtprotobuf v0.6.0",
			1)
	}
	if withTest {
		content = strings.Replace(content,
			"\tgithub.com/ahfuzhang/BaoHuLu v0.4.3",
			"\tgithub.com/ahfuzhang/BaoHuLu v0.4.3\n\tgithub.com/bytedance/sonic v1.15.1",
			1)
		content += sonicIndirectDeps
	}
	return content
}
