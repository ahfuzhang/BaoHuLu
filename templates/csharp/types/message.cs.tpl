{{with csModel .}}
{{template "CsFileHeader" .}}
{{with .Msg}}
{{template "CsTagsBlock" .}}
{{- $goName := .GoName}}
// ============================================================================
// Readonly{{$goName}} — immutable view over decoded protobuf bytes
// ============================================================================

public partial struct Readonly{{$goName}} : IResettable, IDecoder
{
{{- range .Fields}}
    [JsonPropertyName("{{.JsonName}}")]
    public {{.EffReadonlyType}} {{.Name}};
{{- end}}
}
public class Readonly{{$goName}}Wrapper
{
    public Readonly{{$goName}} Value;
}
{{- $goName := .GoName}}
// ============================================================================
// {{$goName}} — mutable struct for building / serialising
// ============================================================================

public partial struct {{$goName}} : IResettable, IEncoder
{
{{- range .Fields}}
    [JsonPropertyName("{{.JsonName}}")]
    public {{.EffWriterType}} {{.Name}};
{{- end}}
}
public class {{$goName}}Wrapper
{
    public {{$goName}} Value;
}
{{end}}{{end}}
