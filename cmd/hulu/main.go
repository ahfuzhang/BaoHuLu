package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/ahfuzhang/BaoHuLu/internal/codegen"
	"github.com/ahfuzhang/BaoHuLu/templates"
)

const usage = `hulu -src=<file.proto> -template.dir=<bundle> -out=<directory>

Always checks before generating. One source, one bundle, one output root per invocation.
  -src           Required input .proto file
  -template.dir  Required external directory or builtin:<name>
  -out           Required output root
  help, -h, -help, --help  Show help
  --version      Show version

Built-in bundles: golang, csharp, csharp/QiWa.rpc
File-header extensions (before syntax):
  // @gen_test_code=on/off       default off
  // @gen_benchmark_code=on/off  default off
  // @gen_project_file=on/off    default on

Migration: remove generate/tu; check/xi have no check-only replacement.
Replace -go_out/-csharp_out with -template.dir=builtin:golang/builtin:csharp and -out.
Replace test/benchmark/project flags with the file-header extensions above.
No -dry-run or -types.manifest. Use a fresh directory for the first migration.
`

var Version = "dev"

func init() {
	if Version != "" && Version != "dev" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}
	if Version == "" {
		Version = "dev"
	}
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "hulu:", err)
		os.Exit(1)
	}
}

func run(arguments []string, stdout, stderr io.Writer) error {
	if len(arguments) == 1 {
		switch arguments[0] {
		case "help", "-h", "-help", "--help":
			_, err := io.WriteString(stdout, usage)
			return err
		case "--version":
			_, err := fmt.Fprintln(stdout, Version)
			return err
		}
	}
	if len(arguments) == 0 {
		return fmt.Errorf("missing generation parameters\n%s", usage)
	}
	for index, argument := range arguments {
		if index != 0 && !strings.HasPrefix(argument, "-") {
			continue
		}
		if hint := migrationHint(argument); hint != "" {
			return fmt.Errorf("%s: %s", argument, hint)
		}
	}
	flags := flag.NewFlagSet("hulu", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { fmt.Fprint(stderr, usage) }
	options := codegen.Options{Version: Version}
	flags.StringVar(&options.Source, "src", "", "input .proto file")
	flags.StringVar(&options.TemplateDir, "template.dir", "", "template bundle directory")
	flags.StringVar(&options.OutputDir, "out", "", "output root directory")
	seen := map[string]bool{}
	for _, argument := range arguments {
		if !strings.HasPrefix(argument, "-") {
			continue
		}
		name := strings.SplitN(strings.TrimLeft(argument, "-"), "=", 2)[0]
		if flags.Lookup(name) != nil {
			if seen[name] {
				return fmt.Errorf("-%s may only be specified once", name)
			}
			seen[name] = true
		}
	}
	if err := flags.Parse(arguments); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s; operation subcommands are not supported", strings.Join(flags.Args(), " "))
	}
	if options.Source == "" || options.TemplateDir == "" || options.OutputDir == "" {
		return fmt.Errorf("-src, -template.dir and -out are required")
	}
	resources, err := templates.Open(options.TemplateDir)
	if err != nil {
		return err
	}
	result, err := codegen.Generate(options, resources)
	if err != nil {
		return err
	}
	for _, warning := range result.Warnings {
		fmt.Fprintln(stderr, "warning:", warning)
	}
	fmt.Fprintf(stdout, "generated %s (%d written, %d removed)\n", options.OutputDir, len(result.Written), len(result.Removed))
	return nil
}

func migrationHint(argument string) string {
	switch argument {
	case "generate", "tu":
		return "remove the subcommand; use hulu -src=<file> -template.dir=<bundle> -out=<dir>"
	case "check", "xi":
		return "check-only mode was removed; generation checks automatically; this invocation will not generate"
	}
	if !strings.HasPrefix(argument, "-") {
		return ""
	}
	name := strings.SplitN(strings.TrimLeft(argument, "-"), "=", 2)[0]
	switch name {
	case "go_out":
		return "use -template.dir=builtin:golang -out=<dir>"
	case "csharp_out":
		return "use -template.dir=builtin:csharp -out=<dir>"
	case "go_out.with.test", "csharp_out.with.test", "with.test":
		return "use // @gen_test_code=on/off before syntax in the proto (default off)"
	case "go_out.with.bench", "csharp_out.with.bench", "with.bench":
		return "use // @gen_benchmark_code=on/off before syntax in the proto (default off)"
	case "with.project":
		return "use // @gen_project_file=on/off before syntax in the proto (default on)"
	case "no.csproj.file":
		return "use // @gen_project_file=off; this disables all new project files, including go.mod and test/benchmark scaffolding"
	case "src.csharp_template.dir":
		return "use -template.dir=<bundle>"
	case "dst.csharp_template.out_dir":
		return "use -out=<dir>"
	}
	return ""
}
