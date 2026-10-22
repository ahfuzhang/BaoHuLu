{{define "fn/csScalarTypes"}}{{return (dict "double" "double" "float" "float" "int32" "int" "int64" "long" "uint32" "uint" "uint64" "ulong" "sint32" "int" "sint64" "long" "fixed32" "uint" "fixed64" "ulong" "sfixed32" "int" "sfixed64" "long" "bool" "bool" "string" "string" "bytes" "byte[]")}}{{end}}
{{define "fn/csType"}}{{- $name := index .Args 0 -}}{{or (get (csScalarTypes) $name) $name}}{{end}}
{{define "fn/snakeToCamel"}}{{- range split (index .Args 0) "_" -}}{{upperFirst .}}{{- end -}}{{end}}
{{define "fn/goTypeName"}}{{index .Args 0}}{{end}}
{{define "fn/csField"}}
{{- $field := index .Args 0 -}}{{- $message := index .Args 1 -}}
{{- $name := snakeToCamel $field.Name -}}{{- $json := $field.Name -}}{{- $yaml := "" -}}
{{- range $field.Annotations -}}
{{- if eq .Name "varname" -}}{{- $name = .Value -}}{{- else if eq .Name "jsonname" -}}{{- $json = .Value -}}{{- else if eq .Name "yamlname" -}}{{- $yaml = .Value -}}{{- end -}}{{- end -}}
{{- $type := $field.Type.RawName -}}
{{- if $field.Type.Message -}}{{- $type = $field.Type.Message.Name -}}{{- else if $field.Type.Enum -}}{{- $type = $field.Type.Enum.Name -}}{{- end -}}
{{- $isMsg := eq $field.Type.Kind "message" "external" -}}{{- $isEnum := eq $field.Type.Kind "enum" -}}
{{- $elem := csType $type -}}{{- $readElem := $elem -}}{{- if $isMsg -}}{{- $readElem = concat "Readonly" $elem -}}{{- end -}}
{{- $writer := $elem -}}{{- $reader := $readElem -}}{{- $decimal := gt $field.DecimalRound 0 -}}
{{- $mapKey := "" -}}{{- $mapVal := "" -}}{{- $mapKeyCS := "" -}}{{- $mapValCS := "" -}}{{- $readMapValCS := "" -}}{{- $mapMsg := false -}}{{- $mapEnum := false -}}
{{- if $field.Map -}}
{{- $mapKey = $field.MapKey.RawName -}}{{- $mapVal = $type -}}{{- $mapMsg = $isMsg -}}{{- $mapEnum = $isEnum -}}
{{- $mapKeyCS = csType $mapKey -}}{{- $mapValCS = $elem -}}{{- $readMapValCS = $readElem -}}
{{- $writer = printf "Dictionary<%s, %s>" $mapKeyCS $elem -}}{{- $reader = printf "Dictionary<%s, %s>" $mapKeyCS $readElem -}}
{{- $type = "map" -}}{{- $isMsg = false -}}{{- $isEnum = false -}}{{- $elem = "map" -}}{{- $readElem = "Readonlymap" -}}
{{- else if $field.Repeated -}}{{- $writer = printf "List<%s>" $elem -}}{{- $reader = printf "List<%s>" $readElem -}}
{{- else if $decimal -}}{{- $writer = "decimal" -}}{{- $reader = "decimal" -}}{{- end -}}
{{- $directWrapper := and $isMsg (not $field.Map) (not $field.Repeated) $field.Recursive -}}
{{- $effWriter := $writer -}}{{- $effReader := $reader -}}{{- $wrapMap := "" -}}{{- $wrapReadMap := "" -}}
{{- if $mapMsg -}}
{{- $wrapMap = concat $mapValCS "Wrapper" -}}{{- $wrapReadMap = concat $readMapValCS "Wrapper" -}}
{{- $effWriter = printf "Dictionary<%s, %s>" $mapKeyCS $wrapMap -}}{{- $effReader = printf "Dictionary<%s, %s>" $mapKeyCS $wrapReadMap -}}
{{- else if $directWrapper -}}{{- $effWriter = concat $writer "Wrapper" -}}{{- $effReader = concat $reader "Wrapper" -}}{{- end -}}
{{- $inline := "" -}}{{- if $field.InlineComments -}}{{- $inline = index $field.InlineComments 0 -}}{{- end -}}
{{- return (dict "Name" $name "JsonName" $json "Number" $field.Number "Comment" $field.Comments "CommentLineNums" $field.CommentLines "InlineComment" $inline "InlineCommentLineNum" $field.InlineCommentLine "YamlName" $yaml "IsMap" $field.Map "Map" $field.Map "IsRepeated" $field.Repeated "Repeated" $field.Repeated "IsMsg" $isMsg "IsEnum" $isEnum "IsString" (eq $type "string") "IsBytes" (eq $type "bytes") "IsBool" (eq $type "bool") "IsSint32" (eq $type "sint32") "IsSint64" (eq $type "sint64") "IsFixed32" (eq $type "float" "fixed32" "sfixed32") "IsFixed64" (and (not $decimal) (eq $type "double" "fixed64" "sfixed64")) "IsPackable" (and $field.Repeated (has (csScalarTypes) $type) (not (eq $type "string" "bytes"))) "IsDecimal" $decimal "DecimalRound" $field.DecimalRound "WriterType" $writer "ReadonlyType" $reader "LocalType" $reader "ElemTypeCS" $elem "ReadonlyElemTypeCS" $readElem "MapKeyCS" $mapKeyCS "MapValCS" $mapValCS "ReadonlyMapValCS" $readMapValCS "MapValIsMsg" $mapMsg "MapValIsEnum" $mapEnum "ElemIsMsg" (and $field.Repeated $isMsg) "MapKey" $mapKey "MapVal" $mapVal "Type" $type "IsRecursive" $field.Recursive "UseMapValWrapper" $mapMsg "UseDirectWrapper" $directWrapper "EffWriterType" $effWriter "EffReadonlyType" $effReader "EffLocalType" $effReader "WrapMapValCS" $wrapMap "WrapReadonlyMapValCS" $wrapReadMap "MapValIsSelfRef" (and $mapMsg (eq $mapVal $message.Name))) -}}
{{end}}

{{define "fn/csDimensions"}}
{{- $field := index .Args 0 -}}{{- $layouts := index .Args 1 -}}
{{- $align := 8 -}}{{- $size := 8 -}}{{- $ptr := 0 -}}
{{- if $field.Map -}}{{- $ptr = 8 -}}
{{- else if $field.Repeated -}}{{- $size = 24 -}}{{- $ptr = 8 -}}
{{- else if $field.IsMsg -}}{{- $size = 24 -}}{{- $ptr = 8 -}}
{{- else if eq $field.Type "string" -}}{{- $size = 16 -}}{{- $ptr = 8 -}}
{{- else if eq $field.Type "bytes" -}}{{- $size = 24 -}}{{- $ptr = 8 -}}
{{- else if or $field.IsEnum (eq $field.Type "float" "int32" "uint32" "sint32" "fixed32" "sfixed32") -}}{{- $align = 4 -}}{{- $size = 4 -}}
{{- else if eq $field.Type "bool" -}}{{- $align = 1 -}}{{- $size = 1 -}}{{- end -}}
{{- if and $field.IsMsg (has $layouts $field.Type) -}}{{- $layout := get $layouts $field.Type -}}{{- $size = $layout.Size -}}{{- $ptr = $layout.Ptrdata -}}{{- end -}}
{{- set $field "Align" $align -}}{{- set $field "Size" $size -}}{{- set $field "Ptrdata" $ptr -}}
{{end}}

{{define "fn/csSchemaModel"}}
{{- validateCSharp -}}
{{- $layouts := dict -}}{{- $messages := list -}}{{- $byID := dict -}}{{- $wrappers := dict -}}
{{- range .Analysis.DependencyOrder -}}{{- if .Generate -}}{{- $message := . -}}{{- $fields := list -}}
{{- range .ActiveFields -}}{{- $field := csField . $message -}}{{- csDimensions $field $layouts -}}{{- $fields = append $fields $field -}}
{{- if $field.UseMapValWrapper -}}{{- set $wrappers $field.MapVal true -}}{{- end -}}
{{- if $field.UseDirectWrapper -}}{{- set $wrappers $field.Type true -}}{{- end -}}
{{- end -}}
{{- $layout := layout $fields -}}{{- set $layouts .Name $layout -}}
{{- $model := dict "Name" .Name "GoName" .Name "Comment" .Comments "CommentLineNums" .CommentLines "Fields" $layout.Fields "NeedsWrapper" false "AsMap" .AsMap "AsArray" .AsArray "UrlValues" .URLValues "Yaml" (index $.Analysis.YAML .ID) -}}
{{- $messages = append $messages $model -}}{{- set $byID .ID $model -}}
{{- end -}}{{- end -}}
{{- range $messages -}}{{- set . "NeedsWrapper" (has $wrappers .Name) -}}{{- end -}}
{{- $file := trimSuffix (base .File.Name) (ext .File.Name) -}}{{- $base := snakeToCamel $file -}}
{{- $namespace := or (index .File.Options "csharp_namespace") .File.Package (upperFirst $file) -}}
{{- return (dict "Namespace" $namespace "BaseFileName" $base "FileBase" $file "Messages" $messages "ByID" $byID "Enums" .Schema.Enums) -}}
{{end}}

{{define "fn/csModel"}}
{{- $shared := memo "csSchemaModel" -}}
{{- $byID := $shared.ByID -}}
{{- $result := dict "Namespace" $shared.Namespace "BaseFileName" $shared.BaseFileName "FileBase" $shared.FileBase "Messages" $shared.Messages "Enums" $shared.Enums "Enum" .Enum "Msg" (dict) -}}
{{- if .Message -}}{{- set $result "Msg" (get $byID .Message.ID) -}}{{- end -}}
{{- return $result -}}
{{end}}
