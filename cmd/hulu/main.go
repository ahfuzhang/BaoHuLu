package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ahfuzhang/BaoHuLu/internal/csharp"
	gogen "github.com/ahfuzhang/BaoHuLu/internal/golang"
	"github.com/ahfuzhang/BaoHuLu/internal/protocheck"
	"github.com/ahfuzhang/BaoHuLu/internal/protofile"
)

const usage = `hulu <command> [flags]

Commands:
  xi/check    Check a .proto file for syntax errors
  tu/generate Generate code from a .proto file
  help        Show this help message

Run "hulu <command> -help" for command-specific flags.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "help", "-help", "--help", "-h":
		fmt.Print(usage)
	case "xi", "check":
		runXi(args)
	case "tu", "generate":
		runTu(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
		os.Exit(1)
	}
}

// runXi checks a .proto file for syntax errors.
func runXi(args []string) {
	fs := flag.NewFlagSet("xi", flag.ExitOnError)
	src := fs.String("src", "", "input .proto file to check")
	fs.Parse(args)

	if *src == "" {
		fmt.Fprintln(os.Stderr, "usage: hulu xi -src=xx.proto")
		os.Exit(1)
	}

	if err := protocheck.Check(*src); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s: OK\n", *src)
}

// writeGoMod creates a go.mod file at modPath for the generated Go package.
// modulePath is derived from the proto file: the full "option go_package" value
// is used when present; otherwise the "package" statement value is used.
// The generated file declares the two runtime dependencies of the generated code.
func writeGoMod(modPath, goPackage, packageName string) error {
	modulePath := goPackage
	if modulePath == "" {
		modulePath = packageName
	}
	if modulePath == "" {
		modulePath = "generated"
	}
	content := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/ahfuzhang/BaoHuLu v0.0.1
	github.com/valyala/fastjson v1.6.4
)
`, modulePath)
	return os.WriteFile(modPath, []byte(content), 0644)
}

// runTu generates Go and/or C# code from a .proto file.
func runTu(args []string) {
	fs := flag.NewFlagSet("tu", flag.ExitOnError)
	src := fs.String("src", "", "input .proto file")
	goOut := fs.String("go_out", "", "output directory for Go code (optional)")
	csOut := fs.String("csharp_out", "", "output directory for C# code (optional)")
	fs.Parse(args)

	if *src == "" {
		fmt.Fprintln(os.Stderr, "usage: hulu tu -src=xx.proto [-go_out=./dir/] [-csharp_out=./dir/]")
		os.Exit(1)
	}
	if *goOut == "" && *csOut == "" {
		fmt.Fprintln(os.Stderr, "hulu tu: at least one of -go_out or -csharp_out must be specified")
		os.Exit(1)
	}

	f, err := os.Open(*src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s: %v\n", *src, err)
		os.Exit(1)
	}
	defer f.Close()

	pg, err := protofile.ParseAndCollect(f, "gen")
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse %s: %v\n", *src, err)
		os.Exit(1)
	}

	base := strings.TrimSuffix(filepath.Base(*src), ".proto")
	goBase := protofile.CamelToSnake(base) // e.g. DemoServer → demo_server
	csBase := protofile.SnakeToCamel(base) // e.g. demo_server → DemoServer

	if *goOut != "" {
		if err := os.MkdirAll(*goOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", *goOut, err)
			os.Exit(1)
		}
		outPath := filepath.Join(*goOut, goBase+".go")
		outF, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", outPath, err)
			os.Exit(1)
		}
		defer outF.Close()

		if err := gogen.NewGenerator(pg).Render(outF); err != nil {
			fmt.Fprintf(os.Stderr, "render: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated %s\n", outPath)

		modPath := filepath.Join(*goOut, "go.mod")
		if err := writeGoMod(modPath, pg.GoPackage, pg.PackageName); err != nil {
			fmt.Fprintf(os.Stderr, "write go.mod: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated %s\n", modPath)
	}

	if *csOut != "" {
		if err := os.MkdirAll(*csOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", *csOut, err)
			os.Exit(1)
		}
		csPath := filepath.Join(*csOut, csBase+".cs")
		csF, err := os.Create(csPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", csPath, err)
			os.Exit(1)
		}
		defer csF.Close()

		ns := protofile.UpperFirst(base)
		if err := csharp.NewGenerator(pg).RenderCS(csF, ns); err != nil {
			fmt.Fprintf(os.Stderr, "renderCS: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated %s\n", csPath)
	}
}
