package csharp

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/ahfuzhang/BaoHuLu/dependencies/golang/utils"
)

//go:embed GrpcGen.csproj.tpl
var grpcGenProjectTemplate string

//go:embed CSharp.csproj.tpl
var csProjectTemplate string

//go:embed CSharp.Benchmarks.Program.cs.tpl
var csBenchmarkProgramTemplate string

// GrpcGenProjectContent renders the embedded GrpcGen project template.
func GrpcGenProjectContent(protoBase string) string {
	return fmt.Sprintf(grpcGenProjectTemplate, protoBase)
}

// ProjectContent renders the embedded C# project template.
func ProjectContent(assemblyName string) string {
	return fmt.Sprintf(csProjectTemplate, assemblyName, assemblyName)
}

// BenchmarkProgramContent returns the embedded BenchmarkDotNet Program.cs content.
func BenchmarkProgramContent() string {
	return csBenchmarkProgramTemplate
}

// WriteGrpcGenProj writes the GrpcGen project file.
func WriteGrpcGenProj(projPath, protoBase string) error {
	return os.WriteFile(projPath, utils.UnsafeBytesFromString(GrpcGenProjectContent(protoBase)), 0644)
}

// WriteBenchmarkProj writes the benchmark project file.
func WriteBenchmarkProj(projPath, assemblyName, namespace, mainProjFile string) error {
	content := BenchmarkProjectContent(assemblyName, namespace+".Benchmarks", mainProjFile)
	return os.WriteFile(projPath, utils.UnsafeBytesFromString(content), 0644)
}

// WriteBenchmarkProgram writes the BenchmarkDotNet entry point.
func WriteBenchmarkProgram(programPath string) error {
	return os.WriteFile(programPath, utils.UnsafeBytesFromString(BenchmarkProgramContent()), 0644)
}

// WriteProject writes the main generated C# project file.
func WriteProject(projPath, assemblyName string) error {
	return os.WriteFile(projPath, utils.UnsafeBytesFromString(ProjectContent(assemblyName)), 0644)
}
