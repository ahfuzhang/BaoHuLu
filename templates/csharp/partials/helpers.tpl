{{define "fn/primitiveCSLit"}}{{- $value := index .Args 0 -}}{{- $values := dict "bool" "true" "string" "\"hello\"" "byte[]" "new byte[] { 0x01, 0x02, 0x03 }" "float" "1.5f" "double" "1.5" "decimal" "1.12345m" "long" "42L" "ulong" "42UL" "uint" "42U" "int" "42" -}}{{- if has $values $value -}}{{get $values $value}}{{- else -}}{{""}}{{- end -}}{{end}}
{{define "fn/benchCsMapValLit"}}{{- $value := index .Args 0 -}}{{- $values := dict "bool" "i % 2 == 0" "string" "LargeString" "byte[]" "new byte[] { (byte)(i & 0xFF) }" "float" "(float)i * 0.5f" "double" "(double)i * 0.5" "long" "(long)i" "ulong" "(ulong)i" "uint" "(uint)i" "int" "i" -}}{{- if $value.MapValIsMsg -}}{{printf "new %s { Value = Build%s() }" $value.WrapMapValCS $value.MapValCS}}{{- else -}}{{- if has $values $value.MapValCS -}}{{get $values $value.MapValCS}}{{- else -}}{{printf "(%s)1" $value.MapValCS}}{{- end -}}{{- end -}}{{end}}
{{define "fn/benchCsFixedMapValLit"}}{{- $value := index .Args 0 -}}{{- $values := dict "bool" "true" "string" "\"value\"" "byte[]" "new byte[] { 0x01 }" "float" "1.5f" "double" "1.5" "long" "42L" "ulong" "42UL" "uint" "42U" "int" "42" -}}{{- if $value.MapValIsMsg -}}{{printf "new %s { Value = Build%s() }" $value.WrapMapValCS $value.MapValCS}}{{- else -}}{{- if has $values $value.MapValCS -}}{{get $values $value.MapValCS}}{{- else -}}{{printf "(%s)1" $value.MapValCS}}{{- end -}}{{- end -}}{{end}}
{{define "fn/benchCsElemLit"}}{{- $value := index .Args 0 -}}{{- $values := dict "bool" "i % 2 == 0" "string" "LargeString" "byte[]" "new byte[] { (byte)(i & 0xFF) }" "float" "(float)i * 0.5f" "double" "(double)i * 0.5" "long" "(long)i" "ulong" "(ulong)i" "uint" "(uint)i" "int" "i" -}}{{- if $value.ElemIsMsg -}}{{printf "Build%s()" $value.ElemTypeCS}}{{- else -}}{{- if has $values $value.ElemTypeCS -}}{{get $values $value.ElemTypeCS}}{{- else -}}{{printf "(%s)1" $value.ElemTypeCS}}{{- end -}}{{- end -}}{{end}}
{{define "fn/benchCsScalarLit"}}{{- $value := index .Args 0 -}}{{- $values := dict "bool" "true" "double" "1.5" "float" "1.5f" "sint32" "-1" "sfixed32" "-1" "sint64" "-1L" "sfixed64" "-1L" "int64" "1L" "uint32" "1U" "fixed32" "1U" "uint64" "1UL" "fixed64" "1UL" -}}{{- if $value.IsEnum -}}{{printf "(%s)1" $value.WriterType}}{{- else -}}{{- if has $values $value.Type -}}{{get $values $value.Type}}{{- else -}}1{{- end -}}{{- end -}}{{end}}
{{define "fn/csDefault"}}{{- $value := index .Args 0 -}}{{- $values := dict "bool" "false" "string" "\"\"" "double" "0.0" "float" "0.0f" "decimal" "0m" -}}{{- if has $values $value -}}{{get $values $value}}{{- else -}}{{if or (hasSuffix $value "[]") (hasPrefix $value "Dictionary<")}}null{{else}}0{{end}}{{- end -}}{{end}}
{{define "fn/csSampleLit"}}{{- $field := index .Args 0 -}}
{{- if $field.IsMap -}}{{- $key := or (primitiveCSLit $field.MapKeyCS) "1" -}}{{- $value := "" -}}
{{- if $field.MapValIsMsg -}}{{- if $field.UseMapValWrapper -}}
{{- if $field.MapValIsSelfRef -}}{{- $value = printf "new %s()" $field.WrapMapValCS -}}
{{- else -}}{{- $value = printf "new %s { Value = %sTests.MakeSample%s() }" $field.WrapMapValCS $field.MapValCS $field.MapValCS -}}{{- end -}}
{{- else -}}{{- $value = printf "%sTests.MakeSample%s()" $field.MapValCS $field.MapValCS -}}{{- end -}}
{{- else -}}{{- $value = or (primitiveCSLit $field.MapValCS) (printf "(%s)1" $field.MapValCS) -}}{{- end -}}
{{printf "new %s { { %s, %s } }" (or $field.EffWriterType $field.WriterType) $key $value}}
{{- else if $field.IsRepeated -}}
{{- $value := or (primitiveCSLit $field.ElemTypeCS) (printf "(%s)1" $field.ElemTypeCS) -}}
{{- if $field.ElemIsMsg -}}{{- $value = printf "%sTests.MakeSample%s()" $field.ElemTypeCS $field.ElemTypeCS -}}{{- end -}}
{{printf "new %s { %s }" $field.WriterType $value}}
{{- else if $field.IsMsg -}}{{- if $field.UseDirectWrapper -}}{{printf "new %s()" $field.EffWriterType}}{{- else -}}{{printf "%sTests.MakeSample%s()" $field.WriterType $field.WriterType}}{{- end -}}
{{- else -}}{{or (primitiveCSLit $field.WriterType) (printf "(%s)1" $field.WriterType)}}{{- end -}}{{end}}
{{define "fn/csYamlSampleLit"}}{{- $field := index .Args 0 -}}
{{- if and $field.IsMap (not $field.MapValIsMsg) -}}
{{- $key := or (primitiveCSLit $field.MapKeyCS) "1" -}}{{- if eq $field.MapKeyCS "string" -}}{{- $key = "\"key\\\\slash\"" -}}{{- end -}}
{{- $value := or (primitiveCSLit $field.MapValCS) (printf "(%s)1" $field.MapValCS) -}}{{- if eq $field.MapValCS "string" -}}{{- $value = "\"val:has:colons\"" -}}{{- end -}}
{{printf "new %s { { %s, %s } }" (or $field.EffWriterType $field.WriterType) $key $value}}
{{- else if and $field.IsRepeated (not $field.ElemIsMsg) -}}
{{- $value := or (primitiveCSLit $field.ElemTypeCS) (printf "(%s)1" $field.ElemTypeCS) -}}{{- if eq $field.ElemTypeCS "string" -}}{{- $value = "\"item \\\"quoted\\\"\"" -}}{{- end -}}
{{printf "new %s { %s }" $field.WriterType $value}}
{{- else if and $field.IsString (not $field.IsRepeated) (not $field.IsMap) -}}"hello \"world\" [brk] {brc}\nnewline\ttab #comment"
{{- else -}}{{csSampleLit $field}}{{- end -}}{{end}}
{{define "fn/firstCsStringField"}}{{- $result := dict -}}{{- range index .Args 0 -}}{{- if and (not $result) (not .IsMap) (not .IsRepeated) .IsString -}}{{- $result = . -}}{{- end -}}{{- end -}}{{return $result}}{{end}}
{{define "fn/csDecimalFields"}}{{- $result := list -}}{{- range index .Args 0 -}}{{- if and .IsDecimal (not .IsMap) (not .IsRepeated) -}}{{- $result = append $result . -}}{{- end -}}{{- end -}}{{return $result}}{{end}}
{{define "fn/anyMsgHasYaml"}}{{- $result := false -}}{{- range index .Args 0 -}}{{- if .Yaml -}}{{- $result = true -}}{{- end -}}{{- end -}}{{return $result}}{{end}}
{{define "fn/safeUnknownFieldNum"}}{{- $number := 0 -}}{{- range index .Args 0 -}}{{- if gt .Number $number -}}{{- $number = .Number -}}{{- end -}}{{- end -}}{{return (add $number 100)}}{{end}}
{{define "fn/tagSize"}}{{return (len (tagBytes (index .Args 0) (index .Args 1)))}}{{end}}
{{define "fn/csTagByteArray"}}new byte[] { {{range $index, $value := tagBytes (index .Args 0) (index .Args 1)}}{{if $index}}, {{end}}{{printf "0x%02X" $value}}{{end}} }{{end}}
{{define "fn/benchCsMapFill"}}{{- $field := index .Args 0 -}}
{{- if eq $field.MapKeyCS "bool" -}}{{- $fixed := benchCsFixedMapValLit $field -}}{{printf "dict[true] = %s; dict[false] = %s;" $fixed $fixed}}
{{- else -}}{{- $key := "i" -}}{{- if eq $field.MapKeyCS "string" -}}{{- $key = "i.ToString()" -}}{{- else if eq $field.MapKeyCS "long" "ulong" "uint" -}}{{- $key = printf "(%s)i" $field.MapKeyCS -}}{{- end -}}
{{printf "for (var i = 0; i < 101; i++) { dict[%s] = %s; }" $key (benchCsMapValLit $field)}}{{- end -}}{{end}}
{{define "fn/benchCsSliceFill"}}{{printf "for (var i = 0; i < 101; i++) { lst.Add(%s); }" (benchCsElemLit (index .Args 0))}}{{end}}
{{define "fn/csUrlKeyParse"}}{{- $type := index .Args 0 -}}{{- $name := index .Args 1 -}}
{{- if eq $type "string" -}}var _key = _mk;{{- else if eq $type "bool" -}}var _key = _mk == "true" || _mk == "1";{{- else -}}
{{- if not (eq $type "long" "ulong" "uint") -}}{{- $type = "int" -}}{{- end -}}if (!{{$type}}.TryParse(_mk, out {{$type}} _key)) return global::QiWa.Common.Error.WithLoc(1, "bad key {{$name}}");{{- end -}}{{end}}
{{define "fn/yamlKey"}}{{- $field := index .Args 0 -}}{{or $field.YamlName $field.JsonName}}{{end}}
{{define "fn/csYamlCommentExpr"}}{{- $lines := index .Args 1 -}}{{- $index := index .Args 2 -}}{{- if and (lt $index (len $lines)) (gt (index $lines $index) 0) -}}{{printf "_%s_YamlComments.Line%d" (index .Args 0) (index $lines $index)}}{{- else -}}{{printf "%q" (concat (index .Args 3) "\n")}}{{- end -}}{{end}}
{{define "fn/csYamlInlineCommentExpr"}}{{- $field := index .Args 1 -}}{{- if gt $field.InlineCommentLineNum 0 -}}{{printf "_%s_YamlComments.Line%d" (index .Args 0) $field.InlineCommentLineNum}}{{- else -}}{{printf "%q" (concat $field.InlineComment "\n")}}{{- end -}}{{end}}
{{define "fn/hasAnyCsYamlComments"}}{{- $message := index .Args 0 -}}{{- $has := gt (len $message.CommentLineNums) 0 -}}{{- range $message.Fields -}}{{- if or .CommentLineNums (gt .InlineCommentLineNum 0) -}}{{- $has = true -}}{{- end -}}{{- end -}}{{return $has}}{{end}}
