<Project Sdk="Microsoft.NET.Sdk">

  <PropertyGroup>
    <TargetFramework>net10.0</TargetFramework>
    <Nullable>enable</Nullable>
    <ImplicitUsings>enable</ImplicitUsings>
    <AssemblyName>GrpcGen</AssemblyName>
  </PropertyGroup>

  <ItemGroup>
    <PackageReference Include="Google.Protobuf" Version="3.*" />
    <!-- Grpc.Tools compiles the .proto file; PrivateAssets=all so it is build-only. -->
    <PackageReference Include="Grpc.Tools" Version="2.*">
      <PrivateAssets>all</PrivateAssets>
    </PackageReference>
  </ItemGroup>

  <ItemGroup>
    <!-- GrpcServices="None" skips gRPC service-stub generation; we only need message types. -->
    <Protobuf Include="%s" GrpcServices="None" />
  </ItemGroup>

</Project>
