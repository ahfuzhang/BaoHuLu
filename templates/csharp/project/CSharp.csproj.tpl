{{with csModel .}}
<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <AssemblyName>{{.BaseFileName}}</AssemblyName>
    <RootNamespace>{{.BaseFileName}}</RootNamespace>
  </PropertyGroup>

  <ItemGroup>
    <Compile Remove="Tests/**/*.cs" />
    <Compile Remove="Benchmarks/**/*.cs" />
  </ItemGroup>

  <ItemGroup>
    <PackageReference Include="QiWa.Common" Version="*" />
  </ItemGroup>

</Project>
{{end}}
