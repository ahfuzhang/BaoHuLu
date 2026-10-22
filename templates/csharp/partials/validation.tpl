{{define "fn/validateCSharp"}}
{{- $keywords := split "abstract as base bool break byte case catch char checked class const continue decimal default delegate do double else enum event explicit extern false finally fixed float for foreach goto if implicit in int interface internal is lock long namespace new null object operator out override params private protected public readonly ref return sbyte sealed short sizeof stackalloc static string struct switch this throw true try typeof uint ulong unchecked unsafe ushort using virtual void volatile while" " " -}}
{{- $file := trimSuffix (base .File.Name) (ext .File.Name) -}}
{{- $namespace := or (index .File.Options "csharp_namespace") .File.Package (upperFirst $file) -}}
{{- range split $namespace "." -}}{{- $segment := . -}}{{- if not (matches "^[A-Za-z_][A-Za-z0-9_]*$" .) -}}{{fail (printf "invalid C# namespace %s" $namespace)}}{{- end -}}{{- range $keywords -}}{{- if eq . $segment -}}{{fail (printf "C# namespace contains keyword %s" .)}}{{- end -}}{{- end -}}{{- end -}}
{{- if not (matches "^[A-Za-z_][A-Za-z0-9_]*$" (snakeToCamel $file)) -}}{{fail (printf "C# source basename cannot form helper/project symbols: %s" $file)}}{{- end -}}
{{- $symbols := dict "Array" true "Math" true "Error" true "RentedBuffer" true "JsonSerializer" true "Encoding" true "Convert" true "BitConverter" true "BitOperations" true -}}
{{- range .Schema.Enums -}}
{{- if has $symbols .Name -}}{{fail (printf "%s: duplicate generated C# symbol %s" .Location .Name)}}{{- end -}}{{- set $symbols .Name true -}}
{{- $enum := . -}}{{- range .Values -}}{{- $value := . -}}{{- if eq .Name $enum.Name -}}{{fail (printf "%s: C# enum member conflicts with its type" .Location)}}{{- end -}}{{- range $keywords -}}{{- if eq . $value.Name -}}{{fail (printf "%s: C# enum member is a keyword: %s" $value.Location .)}}{{- end -}}{{- end -}}{{- end -}}
{{- end -}}
{{- range .Schema.Messages -}}{{- if .Generate -}}{{- $message := . -}}
{{- range list .Name (concat "Readonly" .Name) (concat .Name "Wrapper") (concat "Readonly" .Name "Wrapper") (concat .Name "Tags") -}}{{- if has $symbols . -}}{{fail (printf "%s: duplicate generated C# symbol %s" $message.Location .)}}{{- end -}}{{- set $symbols . true -}}{{- end -}}
{{- $fields := dict .Name true (concat "Readonly" .Name) true "Reset" true "Clone" true "ToJSON" true "FromJSON" true "FromJSONReader" true "ToProtobuf" true "FromProtobuf" true "ProtobufSize" true "ToYAML" true "FromYAML" true "ToURLValues" true "FromURLValues" true -}}
{{- range .ActiveFields -}}{{- $field := csField . $message -}}{{- $name := $field.Name -}}
{{- range list $field.JsonName $field.YamlName -}}{{- if matches "[\"\\\\\x00-\x1f]" . -}}{{fail (printf "%s: field alias needs escaping unsupported by the unchanged C# template" $message.Location)}}{{- end -}}{{- end -}}
{{- if or (not (matches "^[A-Za-z_][A-Za-z0-9_]*$" $name)) (has $fields $name) -}}{{fail (printf "%s: invalid or colliding C# field %s" .Location $name)}}{{- end -}}{{- set $fields $name true -}}
{{- range $keywords -}}{{- if eq . $name -}}{{fail (printf "C# field is a keyword: %s" $name)}}{{- end -}}{{- end -}}
{{- end -}}{{- end -}}{{- end -}}
{{- range $keywords -}}{{- if has $symbols . -}}{{fail (printf "C# type is a keyword: %s" .)}}{{- end -}}{{- end -}}
{{end}}
