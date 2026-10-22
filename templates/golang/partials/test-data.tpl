{{define "fn/sampleScalarFallback"}}{{- if index .Args 0 -}}{{printf "%s(1)" (index .Args 0)}}{{- else -}}1{{- end -}}{{end}}
{{define "fn/benchSliceFallback"}}{{- $field := index .Args 0 -}}{{- if $field.IsMsg -}}{{printf "for i := 0; i < 101; i++ { s[i] = benchBuild%s() }" (elemType $field.GoType)}}{{- else -}}for i := 0; i < 101; i++ { s[i] = 1 }{{- end -}}{{end}}
{{define "fn/mapKeyZeroLit"}}{{- $field := index .Args 0 -}}{{- if eq $field.MapKey "string" -}}""{{- else if eq $field.MapKey "bool" -}}false{{- else -}}{{mapKeyGoType $field.MapKey}}(0){{- end -}}{{end}}
{{define "fn/msgValMapGoType"}}{{- $field := index .Args 0 -}}map[{{mapKeyGoType $field.MapKey}}]*{{$field.MapVal}}{{end}}
{{define "fn/mapWriterGoType"}}{{- $field := index .Args 0 -}}{{- if $field.MapValIsMsg -}}{{msgValMapGoType $field}}{{- else -}}{{$field.GoType}}{{- end -}}{{end}}
{{define "fn/sampleKey"}}{{- $type := index .Args 0 -}}{{- if eq $type "string" -}}"k"{{- else if eq $type "bool" -}}true{{- else -}}{{sampleScalar $type ""}}{{- end -}}{{end}}
{{define "fn/sampleLit"}}
{{- $field := index .Args 0 -}}
{{- if $field.Map -}}{{- $key := sampleKey $field.MapKey -}}
{{- if $field.MapValIsMsg -}}{{- if $field.ElemIsRecursive -}}nil{{- else -}}
{{printf "(func() map[%s]*%s { v := makeSample%s(); return map[%s]*%s{%s: &v} }())" (mapKeyGoType $field.MapKey) $field.MapVal $field.MapVal (mapKeyGoType $field.MapKey) $field.MapVal $key}}
{{- end -}}{{- else -}}{{printf "%s{%s: %s}" $field.GoType $key (sampleScalar $field.MapVal $field.MapVal)}}{{- end -}}
{{- else if $field.Repeated -}}{{- if and $field.IsMsg $field.ElemIsRecursive -}}nil{{- else if $field.IsMsg -}}{{printf "%s{makeSample%s()}" $field.GoType (elemType $field.GoType)}}{{- else -}}{{printf "%s{%s}" $field.GoType (sampleScalar $field.Type (elemType $field.GoType))}}{{- end -}}
{{- else if $field.IsMsg -}}{{- if $field.IsRecursive -}}nil{{- else -}}{{printf "makeSample%s()" $field.GoType}}{{- end -}}
{{- else if $field.IsDecimal -}}decimal.MustParse("1.12345"){{- else -}}{{sampleScalar $field.Type $field.GoType}}{{- end -}}
{{end}}
{{define "fn/sampleLitFull"}}
{{- $field := index .Args 0 -}}
{{- if and $field.IsMsg $field.IsRecursive -}}{{printf "(func() *%s { v := makeSample%sBase(); return &v }())" $field.GoType $field.GoType}}
{{- else if and $field.Map $field.MapValIsMsg $field.ElemIsRecursive -}}
{{- if eq $field.MapKey "bool" -}}nil{{- else -}}{{printf "(func() map[%s]*%s { v := makeSample%sBase(); return map[%s]*%s{%s: &v} }())" (mapKeyGoType $field.MapKey) $field.MapVal $field.MapVal (mapKeyGoType $field.MapKey) $field.MapVal (sampleKey $field.MapKey)}}{{- end -}}
{{- else if and $field.Repeated $field.IsMsg $field.ElemIsRecursive -}}{{printf "%s{makeSample%sBase()}" $field.GoType (elemType $field.GoType)}}
{{- else -}}nil{{- end -}}
{{end}}
{{define "fn/benchMapFill"}}
{{- $field := index .Args 0 -}}{{- $type := mapKeyGoType $field.MapKey -}}
{{- if $field.MapValIsMsg -}}
{{- if eq $field.MapKey "string" -}}{{printf "for i := 0; i < 101; i++ { v := benchBuild%s(); m[strconv.Itoa(i)] = &v }" $field.MapVal}}
{{- else if eq $field.MapKey "bool" -}}{{printf "{ v1 := benchBuild%s(); m[false] = &v1; v2 := benchBuild%s(); m[true] = &v2 }" $field.MapVal $field.MapVal}}
{{- else -}}{{printf "for i := %s(0); i < 101; i++ { v := benchBuild%s(); m[i] = &v }" $type $field.MapVal}}{{- end -}}
{{- else -}}{{- $literal := benchScalarMapValLit $field.MapVal -}}
{{- if eq $field.MapKey "string" -}}{{printf "for i := 0; i < 101; i++ { m[strconv.Itoa(i)] = %s }" $literal}}
{{- else if eq $field.MapKey "bool" -}}{{printf "m[false] = %s\n\t\tm[true] = %s" $literal $literal}}
{{- else -}}{{printf "for i := %s(0); i < 101; i++ { m[i] = %s }" $type $literal}}{{- end -}}{{- end -}}
{{end}}
{{define "fn/benchMapFillRecursive"}}
{{- $field := index .Args 0 -}}{{- $key := printf "%s(0)" (mapKeyGoType $field.MapKey) -}}
{{- if eq $field.MapKey "string" -}}{{- $key = "strconv.Itoa(0)" -}}{{- else if eq $field.MapKey "bool" -}}{{- $key = "0" -}}{{- end -}}
{{printf "{ v := benchBuild%sBase(); m[%s] = &v }" $field.MapVal $key}}
{{end}}
{{define "fn/benchSliceFillRecursive"}}{{printf "s[0] = benchBuild%sBase()" (elemType (index .Args 0).GoType)}}{{end}}
{{define "fn/boundaryLits"}}
{{- $type := index .Args 0 -}}{{- $result := list -}}
{{- if eq $type "int32" "sint32" "sfixed32" -}}{{- $result = list (dict "Label" "max" "Lit" "int32(math.MaxInt32)") (dict "Label" "min" "Lit" "int32(math.MinInt32)") -}}
{{- else if eq $type "uint32" "fixed32" -}}{{- $result = list (dict "Label" "max" "Lit" "uint32(math.MaxUint32)") -}}
{{- else if eq $type "int64" "sint64" "sfixed64" -}}{{- $result = list (dict "Label" "max" "Lit" "int64(math.MaxInt64)") (dict "Label" "min" "Lit" "int64(math.MinInt64)") -}}
{{- else if eq $type "uint64" "fixed64" -}}{{- $result = list (dict "Label" "max" "Lit" "uint64(math.MaxUint64)") -}}
{{- else if eq $type "float" -}}{{- $result = list (dict "Label" "max" "Lit" "float32(math.MaxFloat32)") (dict "Label" "neg_max" "Lit" "float32(-math.MaxFloat32)") (dict "Label" "smallest" "Lit" "float32(math.SmallestNonzeroFloat32)") -}}
{{- else if eq $type "double" -}}{{- $result = list (dict "Label" "max" "Lit" "math.MaxFloat64") (dict "Label" "neg_max" "Lit" "-math.MaxFloat64") (dict "Label" "smallest" "Lit" "math.SmallestNonzeroFloat64") -}}
{{- end -}}{{return $result}}
{{end}}
{{define "fn/numericBoundaryCases"}}
{{- $result := list -}}{{- range $field := index .Args 0 -}}
{{- if and (not $field.Map) (not $field.Repeated) (not $field.IsMsg) (not $field.IsRawBuf) (not $field.IsDecimal) -}}
{{- range boundaryLits $field.Type -}}{{- $result = append $result (dict "Label" (concat $field.Name "_" .Label) "FieldName" $field.Name "Lit" .Lit) -}}{{- end -}}
{{- end -}}{{- end -}}{{return $result}}
{{end}}
{{define "fn/hasNumericBoundaryFields"}}{{return (gt (len (numericBoundaryCases (index .Args 0))) 0)}}{{end}}
