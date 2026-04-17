package csharp

import (
	_ "embed"
	"fmt"
)

//go:embed CSharp.Benchmarks.csproj.tpl
var csBenchProjectTemplate string

// BenchmarkProjectContent renders the embedded C# benchmark project template.
func BenchmarkProjectContent(assemblyName, namespace, mainProjFile string) string {
	return fmt.Sprintf(csBenchProjectTemplate, assemblyName, namespace, mainProjFile)
}
