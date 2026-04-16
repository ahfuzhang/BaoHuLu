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
// The generated file declares the runtime dependencies of the generated code.
func writeGoMod(modPath, goPackage, packageName string, withBench bool) error {
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

// writeCsTestProj creates a .csproj file for the generated C# test project.
// mainProjFile is the filename (without path) of the main project file, e.g. "DemoServer.csproj".
func writeCsTestProj(projPath, assemblyName, namespace, mainProjFile string) error {
	content := fmt.Sprintf(`<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <IsPackable>false</IsPackable>
    <AssemblyName>%s</AssemblyName>
    <RootNamespace>%s</RootNamespace>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Microsoft.NET.Test.Sdk" Version="17.*" />
    <PackageReference Include="xunit" Version="2.*" />
    <PackageReference Include="xunit.runner.visualstudio" Version="2.*">
      <IncludeAssets>runtime; build; native; contentfiles; analyzers; buildtransitive</IncludeAssets>
      <PrivateAssets>all</PrivateAssets>
    </PackageReference>
    <PackageReference Include="QiWa.Common" Version="*" />
  </ItemGroup>

  <ItemGroup>
    <ProjectReference Include="..\%s" />
  </ItemGroup>

</Project>
`, assemblyName, namespace, mainProjFile)
	return os.WriteFile(projPath, []byte(content), 0644)
}

// writeCsProj creates a .csproj file for the generated C# project.
func writeCsProj(projPath, assemblyName string) error {
	content := fmt.Sprintf(`<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <AssemblyName>%s</AssemblyName>
    <RootNamespace>%s</RootNamespace>
  </PropertyGroup>

  <ItemGroup>
    <Compile Remove="Tests/**/*.cs" />
  </ItemGroup>

  <ItemGroup>
    <PackageReference Include="QiWa.Common" Version="*" />
  </ItemGroup>

</Project>
`, assemblyName, assemblyName)
	return os.WriteFile(projPath, []byte(content), 0644)
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

		gen := gogen.NewGenerator(pg)
		if err := gen.Render(outF); err != nil {
			fmt.Fprintf(os.Stderr, "render: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated %s\n", outPath)

		if *goOutWithTest {
			testPath := filepath.Join(*goOut, goBase+"_test.go")
			testF, err := os.Create(testPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "create %s: %v\n", testPath, err)
				os.Exit(1)
			}
			defer testF.Close()

			if err := gen.RenderTest(testF); err != nil {
				fmt.Fprintf(os.Stderr, "render test: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", testPath)
		}

		if *goOutWithBench {
			benchPath := filepath.Join(*goOut, goBase+"_timing_test.go")
			benchF, err := os.Create(benchPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "create %s: %v\n", benchPath, err)
				os.Exit(1)
			}
			defer benchF.Close()

			if err := gen.RenderBench(benchF); err != nil {
				fmt.Fprintf(os.Stderr, "render bench: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("generated %s\n", benchPath)
		}

		modPath := filepath.Join(*goOut, "go.mod")
		if err := writeGoMod(modPath, pg.GoPackage, pg.PackageName, *goOutWithBench); err != nil {
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

		ns := pg.CsharpNamespace
		if ns == "" {
			ns = pg.PackageName
		}
		if ns == "" {
			ns = protofile.UpperFirst(base)
		}
		if err := csharp.NewGenerator(pg).RenderCS(csF, ns); err != nil {
			fmt.Fprintf(os.Stderr, "renderCS: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("generated %s\n", csPath)

		projPath := filepath.Join(*csOut, csBase+".csproj")
		if err := writeCsProj(projPath, csBase); err != nil {
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
	}
}
