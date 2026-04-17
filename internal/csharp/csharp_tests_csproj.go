package csharp

import (
	_ "embed"
	"fmt"
)

//go:embed CSharp.Tests.csproj.tpl
var csTestProjectTemplate string

// TestProjectContent renders the embedded C# test project template.
func TestProjectContent(assemblyName, namespace, mainProjFile string) string {
	return fmt.Sprintf(csTestProjectTemplate, assemblyName, namespace, mainProjFile)
}
