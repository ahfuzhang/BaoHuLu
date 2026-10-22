{{- define "CsYamlWriterFile"}}{{template "CsYamlFileHeader" .}}
{{- with .Msg}}
{{- $goName := .GoName}}
{{- $msgCmtNums := .CommentLineNums}}
{{$hasYamlNames := false}}
{{- range .Fields}}{{if .YamlName}}{{$hasYamlNames = true}}{{end}}{{end -}}
{{- if $hasYamlNames}}
// yaml key constants for {{$goName}}.
internal static class _{{$goName}}_YamlKeys
{
{{- range .Fields}}
    internal const string {{.Name}} = "{{yamlKey .}}";
{{- end}}
}
{{end}}
{{- if hasAnyCsYamlComments .}}
// yaml comment constants for {{$goName}}.
internal static class _{{$goName}}_YamlComments
{
{{- range $i, $line := .Comment}}
{{- if gt (index $msgCmtNums $i) 0}}
    internal static ReadOnlySpan<byte> Line{{index $msgCmtNums $i}} => {{printf "%q" (printf "%s\n" $line)}}u8;
{{- end}}
{{- end}}
{{- range .Fields}}
{{- $f := .}}
{{- range $i, $line := .Comment}}
{{- if gt (index $f.CommentLineNums $i) 0}}
    internal static ReadOnlySpan<byte> Line{{index $f.CommentLineNums $i}} => {{printf "%q" (printf "%s\n" $line)}}u8;
{{- end}}
{{- end}}
{{- if gt .InlineCommentLineNum 0}}
    internal static ReadOnlySpan<byte> Line{{.InlineCommentLineNum}} => {{printf "%q" (printf "%s\n" .InlineComment)}}u8;
{{- end}}
{{- end}}
}
{{- end}}
// ============================================================================
// {{$goName}} — ToYAML serialization
// ============================================================================

public partial struct {{$goName}}
{
    public readonly void ToYAML(ref RentedBuffer _dst, int _indent)
    {
        string _sp = _indent > 0 ? new string(' ', _indent) : string.Empty;
        string _sp2 = new string(' ', _indent + 2);
        string _strVal(string s)
        {
            if (string.IsNullOrEmpty(s)) return s;
            foreach (char _c in s)
                if (_c < 0x20 || _c == ':' || _c == '#' || _c == '"' || _c == '\'' || _c == '\\' ||
                    _c == '{' || _c == '}' || _c == '[' || _c == ']')
                    return "\"" + s.Replace("\\", "\\\\").Replace("\"", "\\\"")
                        .Replace("\n", "\\n").Replace("\r", "\\r").Replace("\t", "\\t") + "\"";
            return s;
        }
{{- if .Comment}}
        int _msgCmtMark = _dst.Length;
{{- range $i, $line := .Comment}}
        _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $msgCmtNums $i $line}});
{{- end}}
        int _msgCmtBody = _dst.Length;
{{- end}}
{{- if .AsArray}}
{{- $fa := index .Fields 0}}
        // @AsArray: serialize the single repeated field directly as bare YAML array members (no wrapper key).
        if (this.{{$fa.Name}} != null)
            foreach (var _va{{$fa.Name}} in this.{{$fa.Name}})
            {
                _dst.Append(_sp); _dst.Append("- ");
{{- if $fa.ElemIsMsg}}
                _dst.Append("\n");
                _va{{$fa.Name}}.ToYAML(ref _dst, _indent + 2);
{{- else if eq $fa.Type "string"}}
                _dst.Append(_strVal(_va{{$fa.Name}} ?? string.Empty)); _dst.Append("\n");
{{- else if eq $fa.Type "bytes"}}
                _dst.Append(Convert.ToBase64String(_va{{$fa.Name}} ?? System.Array.Empty<byte>())); _dst.Append("\n");
{{- else if eq $fa.Type "bool"}}
                _dst.Append(_va{{$fa.Name}} ? "true" : "false"); _dst.Append("\n");
{{- else if eq $fa.Type "double"}}
                _dst.Append(_va{{$fa.Name}}.ToString("R", CultureInfo.InvariantCulture)); _dst.Append("\n");
{{- else if eq $fa.Type "float"}}
                _dst.Append(_va{{$fa.Name}}.ToString("R", CultureInfo.InvariantCulture)); _dst.Append("\n");
{{- else}}
                _dst.Append(_va{{$fa.Name}}.ToString()); _dst.Append("\n");
{{- end}}
            }
{{- else}}
{{range .Fields}}
{{- $cmt := .Comment}}
{{- $cmtNums := .CommentLineNums}}
{{- if .IsMap}}
        if (this.{{.Name}} != null && this.{{.Name}}.Count > 0)
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":  "{{else}}"{{yamlKey .}}:  "{{end}}); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":\n"{{else}}"{{yamlKey .}}:\n"{{end}});
{{- end}}
            foreach (var _kv{{.Name}} in this.{{.Name}})
            {
                _dst.Append(_sp2);
{{- if eq .MapKeyCS "string"}}
                _dst.Append(_strVal(_kv{{.Name}}.Key ?? string.Empty));
{{- else if eq .MapKeyCS "bool"}}
                _dst.Append(_kv{{.Name}}.Key ? "true" : "false");
{{- else}}
                _dst.Append(_kv{{.Name}}.Key.ToString());
{{- end}}
                _dst.Append(": ");
{{- if .MapValIsMsg}}
                _dst.Append("\n");
{{- if .UseMapValWrapper}}
                if (_kv{{.Name}}.Value != null) _kv{{.Name}}.Value.Value.ToYAML(ref _dst, _indent + 4);
{{- else}}
                _kv{{.Name}}.Value.ToYAML(ref _dst, _indent + 4);
{{- end}}
{{- else if eq .MapVal "string"}}
                _dst.Append(_strVal(_kv{{.Name}}.Value ?? string.Empty)); _dst.Append("\n");
{{- else if eq .MapVal "bytes"}}
                _dst.Append(Convert.ToBase64String(_kv{{.Name}}.Value ?? System.Array.Empty<byte>())); _dst.Append("\n");
{{- else if eq .MapVal "bool"}}
                _dst.Append(_kv{{.Name}}.Value ? "true" : "false"); _dst.Append("\n");
{{- else if eq .MapVal "double"}}
                _dst.Append(_kv{{.Name}}.Value.ToString("R", CultureInfo.InvariantCulture)); _dst.Append("\n");
{{- else if eq .MapVal "float"}}
                _dst.Append(_kv{{.Name}}.Value.ToString("R", CultureInfo.InvariantCulture)); _dst.Append("\n");
{{- else if .MapValIsEnum}}
                _dst.Append(((int)_kv{{.Name}}.Value).ToString()); _dst.Append("\n");
{{- else}}
                _dst.Append(_kv{{.Name}}.Value.ToString()); _dst.Append("\n");
{{- end}}
            }
        }
{{- else if .IsRepeated}}
        if (this.{{.Name}} != null && this.{{.Name}}.Count > 0)
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":  "{{else}}"{{yamlKey .}}:  "{{end}}); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":\n"{{else}}"{{yamlKey .}}:\n"{{end}});
{{- end}}
            foreach (var _v{{.Name}} in this.{{.Name}})
            {
                _dst.Append(_sp2); _dst.Append("- ");
{{- if .ElemIsMsg}}
                _dst.Append("\n");
                _v{{.Name}}.ToYAML(ref _dst, _indent + 4);
{{- else if eq .Type "string"}}
                _dst.Append(_strVal(_v{{.Name}} ?? string.Empty)); _dst.Append("\n");
{{- else if eq .Type "bytes"}}
                _dst.Append(Convert.ToBase64String(_v{{.Name}} ?? System.Array.Empty<byte>())); _dst.Append("\n");
{{- else if eq .Type "bool"}}
                _dst.Append(_v{{.Name}} ? "true" : "false"); _dst.Append("\n");
{{- else if eq .Type "double"}}
                _dst.Append(_v{{.Name}}.ToString("R", CultureInfo.InvariantCulture)); _dst.Append("\n");
{{- else if eq .Type "float"}}
                _dst.Append(_v{{.Name}}.ToString("R", CultureInfo.InvariantCulture)); _dst.Append("\n");
{{- else}}
                _dst.Append(_v{{.Name}}.ToString()); _dst.Append("\n");
{{- end}}
            }
        }
{{- else if .IsMsg}}
{{- if .UseDirectWrapper}}
        if (this.{{.Name}} != null)
        {
            int _mark{{.Name}} = _dst.Length;
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":  "{{else}}"{{yamlKey .}}:  "{{end}}); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":\n"{{else}}"{{yamlKey .}}:\n"{{end}});
{{- end}}
            int _hdr{{.Name}} = _dst.Length;
            this.{{.Name}}.Value.ToYAML(ref _dst, _indent + 2);
            if (_dst.Length == _hdr{{.Name}}) _dst.Length = _mark{{.Name}};
        }
{{- else}}
        {
            int _mark{{.Name}} = _dst.Length;
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":  "{{else}}"{{yamlKey .}}:  "{{end}}); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ":\n"{{else}}"{{yamlKey .}}:\n"{{end}});
{{- end}}
            int _hdr{{.Name}} = _dst.Length;
            this.{{.Name}}.ToYAML(ref _dst, _indent + 2);
            if (_dst.Length == _hdr{{.Name}}) _dst.Length = _mark{{.Name}};
        }
{{- end}}
{{- else if .IsDecimal}}
        if (this.{{.Name}} != 0m)
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": "{{else}}"{{yamlKey .}}: "{{end}});
            _dst.Append(Math.Round(this.{{.Name}}, {{.DecimalRound}}, MidpointRounding.AwayFromZero).ToString("F{{.DecimalRound}}", CultureInfo.InvariantCulture));
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- else if .IsBool}}
        if (this.{{.Name}})
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": true"{{else}}"{{yamlKey .}}: true"{{end}});
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- else if .IsString}}
        if (!string.IsNullOrEmpty(this.{{.Name}}))
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": "{{else}}"{{yamlKey .}}: "{{end}}); _dst.Append(_strVal(this.{{.Name}}));
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- else if .IsBytes}}
        if (this.{{.Name}} != null && this.{{.Name}}.Length > 0)
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": "{{else}}"{{yamlKey .}}: "{{end}}); _dst.Append(Convert.ToBase64String(this.{{.Name}}));
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- else if eq .Type "double"}}
        if (this.{{.Name}} != 0.0)
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": "{{else}}"{{yamlKey .}}: "{{end}}); _dst.Append(this.{{.Name}}.ToString("R", CultureInfo.InvariantCulture));
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- else if eq .Type "float"}}
        if (this.{{.Name}} != 0.0f)
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": "{{else}}"{{yamlKey .}}: "{{end}}); _dst.Append(this.{{.Name}}.ToString("R", CultureInfo.InvariantCulture));
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- else if .IsEnum}}
        if ((int)this.{{.Name}} != 0)
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": "{{else}}"{{yamlKey .}}: "{{end}}); _dst.Append(((int)this.{{.Name}}).ToString());
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- else}}
        if (this.{{.Name}} != {{csDefault .WriterType}})
        {
{{- range $i, $line := $cmt}}
            _dst.Append(_sp); _dst.Append("#"); _dst.Append({{csYamlCommentExpr $goName $cmtNums $i $line}});
{{- end}}
            _dst.Append(_sp); _dst.Append({{if $hasYamlNames}}_{{$goName}}_YamlKeys.{{.Name}} + ": "{{else}}"{{yamlKey .}}: "{{end}}); _dst.Append(this.{{.Name}}.ToString());
{{- if gt .InlineCommentLineNum 0}}
            _dst.Append("  "); _dst.Append("#"); _dst.Append({{csYamlInlineCommentExpr $goName .}});
{{- else}}
            _dst.Append("\n");
{{- end}}
        }
{{- end}}
{{end}}
{{- end}}
{{- if .Comment}}
        if (_dst.Length == _msgCmtBody) _dst.Length = _msgCmtMark;
{{- end}}
    }
}
{{end}}{{end}}
{{with csModel .}}{{template "CsYamlWriterFile" .}}{{end}}
