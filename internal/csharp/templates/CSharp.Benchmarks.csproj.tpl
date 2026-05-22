<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <OutputType>Exe</OutputType>
    <TargetFramework>net10.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <AllowUnsafeBlocks>true</AllowUnsafeBlocks>
    <AssemblyName>%s</AssemblyName>
    <RootNamespace>%s</RootNamespace>
  </PropertyGroup>

  <!-- Exclude the GrpcGen sub-project's sources; they are compiled into GrpcGen.dll separately. -->
  <ItemGroup>
    <Compile Remove="GrpcGen/**/*.cs" />
  </ItemGroup>

  <ItemGroup>
    <PackageReference Include="BenchmarkDotNet" Version="0.14.*" />
    <PackageReference Include="QiWa.Common" Version="*" />
  </ItemGroup>

  <ItemGroup>
    <!-- BaoHuLu-generated types -->
    <ProjectReference Include="..\\%s" />
    <!-- Grpc.Tools-compiled types in the .GrpcProto sub-namespace (no conflict) -->
    <ProjectReference Include="GrpcGen\\GrpcGen.csproj" />
  </ItemGroup>

</Project>
