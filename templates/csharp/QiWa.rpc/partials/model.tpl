{{define "fn/rpcType"}}{{- $type := index .Args 0 -}}{{- if $type.Message -}}{{$type.Message.Name}}{{- else -}}{{$type.RawName}}{{- end -}}{{end}}
{{define "fn/rpcMethod"}}
{{- $method := index .Args 0 -}}
{{- $paths := list -}}{{- range $method.Paths -}}{{- $paths = append $paths (trimSuffix (trimPrefix (printf "%q" .) "\"") "\"") -}}{{- end -}}
{{- return (dict "MethodName" $method.Name "RequestType" (rpcType $method.Request) "ResponseType" (rpcType $method.Response) "Paths" $paths) -}}
{{end}}
{{define "fn/rpcModel"}}
{{- $names := dict -}}
{{- range .Schema.Services -}}{{- range .Methods -}}{{- if has $names .Name -}}{{fail (printf "%s: legacy QiWa template cannot share method name %s across services; needs D16 review" .Location .Name)}}{{- end -}}{{- set $names .Name true -}}{{- end -}}{{- end -}}
{{- $namespace := or (index .File.Options "csharp_namespace") .File.Package (upperFirst (trimSuffix (base .File.Name) (ext .File.Name))) -}}
{{- $methods := list -}}{{- range .Service.Methods -}}{{- $methods = append $methods (rpcMethod .) -}}{{- end -}}
{{- $result := dict "CsharpNamespace" $namespace "ServiceName" .Service.Name "Methods" $methods -}}
{{- if .RPC -}}{{- range $key, $value := rpcMethod .RPC -}}{{- set $result $key $value -}}{{- end -}}{{- end -}}
{{- return $result -}}
{{end}}
