package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils"
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
// The generated file declares the runtime dependencies of the generated code.
func writeGoMod(modPath, goPackage, packageName string) error {
	modulePath := goPackage
	if modulePath == "" {
		modulePath = packageName
	}
	if modulePath == "" {
		modulePath = "generated"
	}
	content := gogen.GoModContent(modulePath)
	return os.WriteFile(modPath, utils.UnsafeBytesFromString(content), 0644)
}

// writeCsTestProj creates a .csproj file for the generated C# test project.
// mainProjFile is the filename (without path) of the main project file, e.g. "DemoServer.csproj".
func writeCsTestProj(projPath, assemblyName, namespace, mainProjFile string) error {
	content := csharp.TestProjectContent(assemblyName, namespace, mainProjFile)
	return os.WriteFile(projPath, utils.UnsafeBytesFromString(content), 0644)
}

// modifyProtoNamespace returns a copy of the proto file content with the
// csharp_namespace option changed from origNS to newNS.
// If no csharp_namespace option exists one is inserted before the first
// message or enum declaration.
func modifyProtoNamespace(data []byte, origNS, newNS string) []byte {
	content := string(data)
	old := `option csharp_namespace = "` + origNS + `"`
	rep := `option csharp_namespace = "` + newNS + `"`
	if strings.Contains(content, old) {
		return utils.UnsafeBytesFromString(strings.Replace(content, old, rep, 1))
	}
	// No existing option — insert one before the first message/enum.
	insert := rep + ";\n"
	for _, kw := range []string{"\nmessage ", "\nenum "} {
		if idx := strings.Index(content, kw); idx >= 0 {
			return utils.UnsafeBytesFromString(content[:idx+1] + insert + content[idx+1:])
		}
	}
	return utils.UnsafeBytesFromString(insert + content)
}

func generateGoOutput(pg *protofile.Generator, goOut, goBase string, withTest, withBench bool) error {
	if err := os.MkdirAll(goOut, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", goOut, err)
	}

	gen := gogen.NewGenerator(pg)
	renderGoFile := func(path string, render func(*os.File) error) error {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		if err := render(f); err != nil {
			f.Close()
			return fmt.Errorf("render %s: %w", path, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", path, err)
		}
		fmt.Printf("generated %s\n", path)
		return nil
	}

	outPath := filepath.Join(goOut, goBase+".go")
	if err := renderGoFile(outPath, gen.Render); err != nil {
		return err
	}

	if withTest {
		testPath := filepath.Join(goOut, goBase+"_test.go")
		if err := renderGoFile(testPath, gen.RenderTest); err != nil {
			return err
		}
	}

	if withBench {
		benchPath := filepath.Join(goOut, goBase+"_timing_test.go")
		if err := renderGoFile(benchPath, gen.RenderBench); err != nil {
			return err
		}
	}

	modPath := filepath.Join(goOut, "go.mod")
	if err := writeGoMod(modPath, pg.GoPackage, pg.PackageName); err != nil {
		return fmt.Errorf("write %s: %w", modPath, err)
	}
	fmt.Printf("generated %s\n", modPath)
	return nil
}

// ─── template generation ──────────────────────────────────────────────────────

type ServiceTplData struct {
	CsharpNamespace string
	ServiceName     string
	Methods         []MethodEntry
	Generator       *protofile.Generator
}

type MethodEntry struct {
	MethodName   string
	RequestType  string
	ResponseType string
	Path         string // @path annotation value; non-empty means an additional HTTP path alias
}

type MethodTplData struct {
	CsharpNamespace string
	ServiceName     string
	MethodName      string
	RequestType     string
	ResponseType    string
	Generator       *protofile.Generator
}

func renderTplToFile(content []byte, data any, outPath string) error {
	t, err := template.New("").Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	if err := t.Execute(f, data); err != nil {
		f.Close()
		return fmt.Errorf("execute template %s: %w", outPath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close %s: %w", outPath, err)
	}
	fmt.Printf("generated %s\n", outPath)
	return nil
}

func processTplFile(pg *protofile.Generator, ns string, content []byte, outBase, dstDir string) error {
	switch {
	case strings.Contains(outBase, "Service"):
		for _, svc := range pg.Services {
			outName := strings.ReplaceAll(outBase, "Service", svc.Name)
			data := ServiceTplData{
				CsharpNamespace: ns,
				ServiceName:     svc.Name,
				Generator:       pg,
			}
			for _, m := range svc.Methods {
				data.Methods = append(data.Methods, MethodEntry{
					MethodName:   m.Name,
					RequestType:  m.RequestType,
					ResponseType: m.ResponseType,
					Path:         m.Path,
				})
			}
			if err := renderTplToFile(content, data, filepath.Join(dstDir, outName)); err != nil {
				return err
			}
		}
	case strings.Contains(outBase, "Method"):
		for _, svc := range pg.Services {
			for _, m := range svc.Methods {
				outName := strings.ReplaceAll(outBase, "Method", svc.Name+m.Name)
				data := MethodTplData{
					CsharpNamespace: ns,
					ServiceName:     svc.Name,
					MethodName:      m.Name,
					RequestType:     m.RequestType,
					ResponseType:    m.ResponseType,
					Generator:       pg,
				}
				if err := renderTplToFile(content, data, filepath.Join(dstDir, outName)); err != nil {
					return err
				}
			}
		}
	default:
		if err := renderTplToFile(content, pg, filepath.Join(dstDir, outBase)); err != nil {
			return err
		}
	}
	return nil
}

func processTemplateDir(pg *protofile.Generator, srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dstDir, err)
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("readdir %s: %w", srcDir, err)
	}
	ns := pg.CsharpNamespace
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".tpl") {
			continue
		}
		tplPath := filepath.Join(srcDir, e.Name())
		content, err := os.ReadFile(tplPath)
		if err != nil {
			return fmt.Errorf("read %s: %w", tplPath, err)
		}
		outBase := strings.TrimSuffix(e.Name(), ".tpl")
		if err := processTplFile(pg, ns, content, outBase, dstDir); err != nil {
			return fmt.Errorf("%s: %w", tplPath, err)
		}
	}
	return nil
}

// runTu generates Go and/or C# code from a .proto file.
func runTu(args []string) {
	fs := flag.NewFlagSet("tu", flag.ExitOnError)
	src := fs.String("src", "", "input .proto file")
	goOut := fs.String("go_out", "", "output directory for Go code (optional)")
	goOutWithTest := fs.Bool("go_out.with.test", false, "also generate _test.go alongside Go output")
	goOutWithBench := fs.Bool("go_out.with.bench", false, "also generate _timing_test.go with benchmark code alongside Go output")
	csOut := fs.String("csharp_out", "", "output directory for C# code (optional)")
	csOutWithTest := fs.Bool("csharp_out.with.test", false, "also generate a Tests/ project with xunit tests alongside C# output")
	csOutWithBench := fs.Bool("csharp_out.with.bench", false, "also generate a Benchmarks/ project with BenchmarkDotNet benchmarks alongside C# output")
	srcTemplateDir := fs.String("src.csharp_template.dir", "", "directory containing .tpl files for C# RPC code generation")
	dstTemplateOutDir := fs.String("dst.csharp_template.out_dir", "", "output directory for files generated from C# .tpl templates")
	fs.Parse(args)

	if *src == "" {
		fmt.Fprintln(os.Stderr, "usage: hulu tu -src=xx.proto [-go_out=./dir/] [-csharp_out=./dir/] [-src.csharp_template.dir=./dir/ -dst.csharp_template.out_dir=./dir/]")
		os.Exit(1)
	}
	if *goOut == "" && *csOut == "" && *srcTemplateDir == "" {
		fmt.Fprintln(os.Stderr, "hulu tu: at least one of -go_out, -csharp_out, or -src.csharp_template.dir must be specified")
		os.Exit(1)
	}
	if *srcTemplateDir != "" && *dstTemplateOutDir == "" {
		fmt.Fprintln(os.Stderr, "hulu tu: -dst.csharp_template.out_dir is required when -src.csharp_template.dir is specified")
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

	if *srcTemplateDir != "" {
		if err := processTemplateDir(pg, *srcTemplateDir, *dstTemplateOutDir); err != nil {
			fmt.Fprintf(os.Stderr, "processTemplateDir: %v\n", err)
			os.Exit(1)
		}
	}

	var goErrCh chan error
	if *goOut != "" {
		goErrCh = make(chan error, 1)
		goOutDir := *goOut
		goWithTest := *goOutWithTest
		goWithBench := *goOutWithBench
		go func() {
			goErrCh <- generateGoOutput(pg, goOutDir, goBase, goWithTest, goWithBench)
		}()
	}

	if *csOut != "" {
		if err := os.MkdirAll(*csOut, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", *csOut, err)
			os.Exit(1)
		}
		ns := pg.CsharpNamespace
		if ns == "" {
			ns = pg.PackageName
		}
		if ns == "" {
			ns = protofile.UpperFirst(base)
		}
		csGen := csharp.NewGenerator(pg)
		if err := csGen.RenderCSFiles(*csOut, csBase, ns); err != nil {
			fmt.Fprintf(os.Stderr, "renderCS: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated %s/*.cs\n", *csOut)

		projPath := filepath.Join(*csOut, csBase+".csproj")
		if err := csharp.WriteProject(projPath, csBase); err != nil {
			fmt.Fprintf(os.Stderr, "write csproj: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated %s\n", projPath)

		if *csOutWithTest {
			testsDir := filepath.Join(*csOut, "Tests")
			if err := os.MkdirAll(testsDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", testsDir, err)
				os.Exit(1)
			}

			testCSPath := filepath.Join(testsDir, csBase+"Tests.cs")
			testCSF, err := os.Create(testCSPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "create %s: %v\n", testCSPath, err)
				os.Exit(1)
			}
			defer testCSF.Close()

			if err := csharp.NewGenerator(pg).RenderCSTest(testCSF, ns); err != nil {
				fmt.Fprintf(os.Stderr, "renderCSTest: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", testCSPath)

			testProjPath := filepath.Join(testsDir, csBase+".Tests.csproj")
			if err := writeCsTestProj(testProjPath, csBase+".Tests", ns+".Tests", csBase+".csproj"); err != nil {
				fmt.Fprintf(os.Stderr, "write test csproj: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", testProjPath)
		}

		if *csOutWithBench {
			benchDir := filepath.Join(*csOut, "Benchmarks")
			if err := os.MkdirAll(benchDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", benchDir, err)
				os.Exit(1)
			}

			// ${Package}.bench.cs — benchmark source
			benchCSPath := filepath.Join(benchDir, csBase+".bench.cs")
			benchCSF, err := os.Create(benchCSPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "create %s: %v\n", benchCSPath, err)
				os.Exit(1)
			}
			defer benchCSF.Close()

			if err := csharp.NewGenerator(pg).RenderCSBench(benchCSF, ns); err != nil {
				fmt.Fprintf(os.Stderr, "renderCSBench: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", benchCSPath)

			// Program.cs — BenchmarkDotNet entry point
			programPath := filepath.Join(benchDir, "Program.cs")
			if err := csharp.WriteBenchmarkProgram(programPath); err != nil {
				fmt.Fprintf(os.Stderr, "write Program.cs: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", programPath)

			// GrpcGen/ sub-project: write a namespace-modified copy of the proto so that
			// Grpc.Tools-compiled types land in ns+".GrpcProto" instead of ns.
			// This avoids type-name collisions with BaoHuLu-generated types in BDN's
			// auto-generated wrapper project (which does not support extern alias).
			grpcGenDir := filepath.Join(benchDir, "GrpcGen")
			if err := os.MkdirAll(grpcGenDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", grpcGenDir, err)
				os.Exit(1)
			}
			protoData, err := os.ReadFile(*src)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read %s: %v\n", *src, err)
				os.Exit(1)
			}
			grpcNS := ns + ".GrpcProto"
			protoBase := filepath.Base(*src)
			grpcProtoPath := filepath.Join(grpcGenDir, protoBase)
			if err := os.WriteFile(grpcProtoPath, modifyProtoNamespace(protoData, ns, grpcNS), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "write %s: %v\n", grpcProtoPath, err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", grpcProtoPath)

			grpcGenProjPath := filepath.Join(grpcGenDir, "GrpcGen.csproj")
			if err := csharp.WriteGrpcGenProj(grpcGenProjPath, protoBase); err != nil {
				fmt.Fprintf(os.Stderr, "write GrpcGen.csproj: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", grpcGenProjPath)

			// ${Package}.bench.csproj
			benchProjPath := filepath.Join(benchDir, csBase+".bench.csproj")
			if err := csharp.WriteBenchmarkProj(benchProjPath, csBase+".bench", ns, csBase+".csproj"); err != nil {
				fmt.Fprintf(os.Stderr, "write bench csproj: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", benchProjPath)
		}
	}

	if goErrCh != nil {
		if err := <-goErrCh; err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
