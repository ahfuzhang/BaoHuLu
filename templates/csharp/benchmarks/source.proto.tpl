{{- $model := csModel . -}}
{{- $source := .File.Source -}}
{{- $namespace := index .File.Options "csharp_namespace" -}}
{{- if $namespace -}}{{- $span := index .File.OptionValues "csharp_namespace" -}}{{- $source = printf "%s%q%s" (slice $source 0 $span.Start) (concat $model.Namespace ".GrpcProto") (slice $source $span.End) -}}{{- else -}}{{- $source = printf "%s\noption csharp_namespace = %q;\n" $source (concat $model.Namespace ".GrpcProto") -}}{{- end -}}
{{$source}}
