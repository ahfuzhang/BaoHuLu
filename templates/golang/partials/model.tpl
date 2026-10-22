{{define "fn/scalarTypes"}}{{return (dict "double" "float64" "float" "float32" "int32" "int32" "int64" "int64" "uint32" "uint32" "uint64" "uint64" "sint32" "int32" "sint64" "int64" "fixed32" "uint32" "fixed64" "uint64" "sfixed32" "int32" "sfixed64" "int64" "bool" "bool" "string" "string" "bytes" "[]byte")}}{{end}}
{{define "fn/mapKeyGoType"}}{{- $name := index .Args 0 -}}{{- or (get (scalarTypes) $name) $name -}}{{end}}
{{define "fn/mapValGoType"}}{{mapKeyGoType (index .Args 0)}}{{end}}
{{define "fn/mapValIsEnum"}}{{- $name := index .Args 0 -}}{{- $found := false -}}{{- range .Schema.Enums -}}{{- if or (eq .Name $name) (eq .FullName $name) -}}{{- $found = true -}}{{- end -}}{{- end -}}{{return $found}}{{end}}
{{define "fn/mapValIsMsg"}}{{- $name := index .Args 0 -}}{{return (and (not (has (scalarTypes) $name)) (not (mapValIsEnum $name)))}}{{end}}
{{define "fn/trimPtr"}}{{trimPrefix (index .Args 0) "*"}}{{end}}
{{define "fn/elemType"}}{{trimPrefix (index .Args 0) "[]"}}{{end}}
{{define "fn/readonlyTypeName"}}Readonly{{index .Args 0}}{{end}}
{{define "fn/readerElemType"}}Readonly{{(index .Args 0).Type}}{{end}}
{{define "fn/snakeToCamel"}}{{- range split (index .Args 0) "_" -}}{{upperFirst .}}{{- end -}}{{end}}
{{define "fn/enumValueGoName"}}{{- range split (index .Args 0) "_" -}}{{upperFirst (lower .)}}{{- end -}}{{end}}
{{define "fn/isPackable"}}{{- $name := index .Args 0 -}}{{return (and (has (scalarTypes) $name) (ne $name "string") (ne $name "bytes"))}}{{end}}
{{define "fn/is8ByteNumeric"}}{{return (eq (index .Args 0) "double" "int64" "uint64" "sint64" "fixed64" "sfixed64")}}{{end}}
{{define "fn/isSliceType"}}{{return (hasPrefix (index .Args 0) "[]")}}{{end}}
{{define "fn/isLargeIntType"}}{{return (eq (index .Args 0) "int64" "uint64" "sint64" "fixed64" "sfixed64")}}{{end}}
{{define "fn/decimalRound"}}{{return (index .Args 0).DecimalRound}}{{end}}
{{define "fn/zeroVal"}}{{- $name := index .Args 0 -}}{{- if eq $name "bool" -}}false{{- else if eq $name "string" -}}""{{- else if eq $name "decimal.Decimal" -}}decimal.Decimal{}{{- else if or (hasPrefix $name "[]") (hasPrefix $name "map[") (hasPrefix $name "*") -}}nil{{- else -}}0{{- end -}}{{end}}
{{define "fn/readerZero"}}{{- $name := index .Args 0 -}}{{- if hasPrefix $name "Readonly" -}}{{$name}}{}{{- else -}}{{zeroVal $name}}{{- end -}}{{end}}
{{define "fn/protoWireTypeInt"}}{{- $name := index .Args 0 -}}{{- $number := 0 -}}{{- if eq $name "double" "fixed64" "sfixed64" -}}{{- $number = 1 -}}{{- else if eq $name "float" "fixed32" "sfixed32" -}}{{- $number = 5 -}}{{- else if eq $name "string" "bytes" -}}{{- $number = 2 -}}{{- end -}}{{return $number}}{{end}}
{{define "fn/wireName"}}{{- index (dict "0" "utils.WireTypeVarint" "1" "utils.WireType64bit" "2" "utils.WireTypeLenDelim" "5" "utils.WireType32bit") (printf "%d" (index .Args 0)) -}}{{end}}
{{define "fn/protoWireType"}}{{wireName (protoWireTypeInt (index .Args 0))}}{{end}}
{{define "fn/tagSize"}}{{return (len (tagBytes (index .Args 0) (index .Args 1)))}}{{end}}
{{define "fn/fieldCommentBlock"}}{{- range index .Args 0 -}}{{printf "\t//%s\n" .}}{{- end -}}{{end}}
{{define "fn/msgCommentBlock"}}{{- range index .Args 0 -}}{{printf "//%s\n" .}}{{- end -}}{{end}}

{{define "fn/goField"}}
{{- $field := index .Args 0 -}}{{- $message := index .Args 1 -}}
{{- $name := snakeToCamel $field.Name -}}{{- $json := $field.Name -}}{{- $yaml := "" -}}{{- $tags := "" -}}
{{- range $field.Annotations -}}
{{- if eq .Name "varname" -}}{{- $name = .Value -}}
{{- else if eq .Name "jsonname" -}}{{- $json = .Value -}}
{{- else if eq .Name "yamlname" -}}{{- $yaml = .Value -}}
{{- else if eq .Name "tag" -}}
{{- $parts := split .Value ":" -}}{{- $key := index $parts 0 -}}{{- $value := trimPrefix .Value (concat $key ":") -}}
{{- $value = trimSuffix (trimPrefix $value "\"") "\"" -}}{{- $tags = printf "%s %s:%q" $tags $key $value -}}
{{- end -}}{{- end -}}
{{- $type := $field.Type.RawName -}}
{{- if $field.Type.Message -}}{{- $type = $field.Type.Message.Name -}}{{- else if $field.Type.Enum -}}{{- $type = $field.Type.Enum.Name -}}{{- end -}}
{{- $isMsg := eq $field.Type.Kind "message" "external" -}}{{- $isEnum := eq $field.Type.Kind "enum" -}}
{{- $goType := mapKeyGoType $type -}}{{- $reader := $goType -}}
{{- if $isMsg -}}{{- $reader = concat "Readonly" $type -}}{{- end -}}
{{- $mapKey := "" -}}{{- $mapVal := "" -}}{{- $mapMsg := false -}}{{- $asArray := false -}}
{{- if $field.Type.Message -}}{{- $asArray = $field.Type.Message.AsArray -}}{{- end -}}
{{- $elemRecursive := and (or $field.Map $field.Repeated) $isMsg (reachable .Analysis.Dependencies $field.Type.SymbolID $message.ID) -}}
{{- if $field.Map -}}
{{- $mapKey = $field.MapKey.RawName -}}{{- $mapVal = $type -}}{{- $mapMsg = $isMsg -}}
{{- $goType = printf "map[%s]%s" (mapKeyGoType $mapKey) $goType -}}
{{- $reader = printf "map[%s]%s" (mapKeyGoType $mapKey) $reader -}}
{{- $type = "map" -}}{{- $isMsg = false -}}{{- $isEnum = false -}}
{{- else if $field.Repeated -}}{{- $goType = concat "[]" $goType -}}{{- $reader = concat "[]" $reader -}}
{{- end -}}
{{- $tag := printf "json:%q" (concat $json ",omitempty") -}}
{{- if $yaml -}}{{- $tag = printf "%s yaml:%q" $tag $yaml -}}{{- end -}}
{{- $tag = printf "%c%s%s%c" 96 $tag $tags 96 -}}
{{- $inline := "" -}}{{- if $field.InlineComments -}}{{- $inline = index $field.InlineComments 0 -}}{{- end -}}
{{- return (dict "Name" $name "Number" $field.Number "JsonName" $json "YamlName" $yaml "Type" $type "GoType" $goType "ReaderType" $reader "Map" $field.Map "Repeated" $field.Repeated "MapKey" $mapKey "MapVal" $mapVal "IsMsg" $isMsg "IsEnum" $isEnum "IsRecursive" $field.Recursive "ElemIsRecursive" $elemRecursive "MapValIsMsg" $mapMsg "MapValMsgIsAsArray" (and $mapMsg $asArray) "MsgIsAsArray" (and $isMsg $asArray) "IsDecimal" (gt $field.DecimalRound 0) "DecimalRound" $field.DecimalRound "IsRawBuf" false "WireType" (wireName $field.WireType) "StructTag" $tag "Comment" $field.Comments "CommentLineNums" $field.CommentLines "InlineComment" $inline "InlineCommentLineNum" $field.InlineCommentLine) -}}
{{end}}

{{define "fn/goDimensions"}}
{{- $field := index .Args 0 -}}{{- $layouts := index .Args 1 -}}
{{- $align := 8 -}}{{- $size := 8 -}}{{- $ptr := 0 -}}
{{- if $field.Map -}}{{- $ptr = 8 -}}
{{- else if $field.Repeated -}}{{- $size = 24 -}}{{- $ptr = 8 -}}
{{- else if $field.IsMsg -}}{{- $size = 24 -}}{{- $ptr = 8 -}}
{{- else if eq $field.Type "string" -}}{{- $size = 16 -}}{{- $ptr = 8 -}}
{{- else if eq $field.Type "bytes" -}}{{- $size = 24 -}}{{- $ptr = 8 -}}
{{- else if or $field.IsEnum (eq $field.Type "float" "int32" "uint32" "sint32" "fixed32" "sfixed32") -}}{{- $align = 4 -}}{{- $size = 4 -}}
{{- else if eq $field.Type "bool" -}}{{- $align = 1 -}}{{- $size = 1 -}}{{- end -}}
{{- if and $field.IsMsg $field.IsRecursive -}}{{- $size = 8 -}}{{- $ptr = 8 -}}
{{- else if and $field.IsMsg (has $layouts $field.Type) -}}{{- $layout := get $layouts $field.Type -}}{{- $size = $layout.Size -}}{{- $ptr = $layout.Ptrdata -}}{{- end -}}
{{- set $field "Align" $align -}}{{- set $field "Size" $size -}}{{- set $field "Ptrdata" $ptr -}}
{{end}}

{{define "fn/goSchemaModel"}}
{{- validateGo -}}
{{- $writerLayouts := dict -}}{{- $readerLayouts := dict -}}{{- $messages := dict -}}{{- $decimal := false -}}
{{- range .Analysis.DependencyOrder -}}
{{- if .Generate -}}
{{- $message := . -}}{{- $fields := list -}}{{- $readerFields := list -}}
{{- range .ActiveFields -}}
{{- $field := goField . $message -}}{{- goDimensions $field $writerLayouts -}}{{- $fields = append $fields $field -}}
{{- $readerField := goField . $message -}}{{- goDimensions $readerField $readerLayouts -}}{{- $readerFields = append $readerFields $readerField -}}
{{- if $field.IsDecimal -}}{{- $decimal = true -}}{{- end -}}
{{- end -}}
{{- $writerLayout := layout $fields -}}{{- set $writerLayouts .Name $writerLayout -}}
{{- $raw := dict "Name" "rawBufferLen" "Number" 0 "Type" "int64" "GoType" "int" "ReaderType" "int" "IsMsg" false "IsEnum" false "Map" false "Repeated" false "IsRawBuf" true -}}
{{- goDimensions $raw $readerLayouts -}}{{- $readerLayout := layout (append $readerFields $raw) -}}{{- set $readerLayouts .Name $readerLayout -}}
{{- set $messages .ID (dict "Name" .Name "GoName" .Name "Comment" .Comments "CommentLineNums" .CommentLines "Fields" $writerLayout.Fields "ReverseFields" (reverse $writerLayout.Fields) "ReaderFields" $readerLayout.Fields "AsMap" .AsMap "AsArray" .AsArray "Yaml" .YAML "UrlValues" .URLValues) -}}
{{- end -}}{{- end -}}
{{- $module := or (index .File.Options "go_package") .File.Package "generated" -}}
{{- $package := $module -}}{{- $parts := split $module "/" -}}{{- $package = index $parts (sub (len $parts) 1) -}}
{{- if contains $module ";" -}}{{- $parts = split $module ";" -}}{{- $module = index $parts 0 -}}{{- $package = index $parts 1 -}}{{- end -}}
{{- return (dict "Package" $package "Messages" $messages "Project" (dict "ModulePath" $module "Version" .Options.Version "WithTest" .File.GenTestCode.Value "WithBench" .File.GenBenchmarkCode.Value "WithDecimal" $decimal)) -}}
{{end}}

{{define "fn/model"}}
{{- $shared := memo "goSchemaModel" -}}
{{- $messages := $shared.Messages -}}
{{- $result := dict "Package" $shared.Package "Enum" .Enum "Msg" (dict) "Project" $shared.Project -}}
{{- if .Message -}}{{- set $result "Msg" (get $messages .Message.ID) -}}{{- end -}}
{{- return $result -}}
{{end}}
