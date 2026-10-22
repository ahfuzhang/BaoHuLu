{{define "fn/validateGo"}}
{{- $keywords := split "break default func interface select case defer go map struct chan else goto package switch const fallthrough if range type continue for import return var" " " -}}
{{- $module := or (index .File.Options "go_package") .File.Package "generated" -}}
{{- $parts := split $module "/" -}}{{- $package := index $parts (sub (len $parts) 1) -}}
{{- if contains $module ";" -}}{{- $parts = split $module ";" -}}{{- if ne (len $parts) 2 -}}{{fail "invalid go_package alias"}}{{- end -}}{{- $module = index $parts 0 -}}{{- $package = index $parts 1 -}}{{- end -}}
{{- if or (not (matches "^[A-Za-z_][A-Za-z0-9_]*$" $package)) (eq $package "_") (not (matches "^[A-Za-z0-9._~/-]+$" $module)) -}}{{fail "invalid Go package or module path"}}{{- end -}}
{{- range $keywords -}}{{- if eq . $package -}}{{fail (printf "Go package is a keyword: %s" $package)}}{{- end -}}{{- end -}}
{{- $symbols := dict -}}
{{- range .Schema.Enums -}}
{{- validateGoFilename .Name .Location -}}
{{- if has $symbols .Name -}}{{fail (printf "%s: duplicate generated symbol %s" .Location .Name)}}{{- end -}}{{- set $symbols .Name true -}}
{{- range .Values -}}{{- $name := enumValueGoName .Name -}}{{- if has $symbols $name -}}{{fail (printf "%s: duplicate enum symbol %s" .Location $name)}}{{- end -}}{{- set $symbols $name true -}}{{- end -}}
{{- end -}}
{{- range .Schema.Messages -}}{{- if .Generate -}}{{- $message := . -}}
{{- validateGoFilename .Name .Location -}}
{{- range list .Name (concat "Readonly" .Name) -}}{{- if has $symbols . -}}{{fail (printf "%s: duplicate generated symbol %s" $message.Location .)}}{{- end -}}{{- set $symbols . true -}}{{- end -}}
{{- $fields := dict "Reset" true "Clone" true "ToJSON" true "FromJSON" true "ToProtobuf" true "FromProtobuf" true "ProtobufSize" true "ToProtobufByAppend" true "ToProtobufVT" true "FromProtobufWithArena" true "ToYAML" true "FromYAML" true "ToURLValues" true "FromURLValues" true "arena" true "parser" true "strArena" true "rawBufferLen" true -}}
{{- range .ActiveFields -}}{{- $field := goField . $message -}}{{- $name := $field.Name -}}
{{- range list $field.JsonName $field.YamlName -}}{{- if matches "[\"\\\\\x00-\x1f\x60]" . -}}{{fail (printf "%s: field alias needs escaping unsupported by the unchanged Go template" $message.Location)}}{{- end -}}{{- end -}}
{{- range .Annotations -}}{{- if and (eq .Name "tag") (matches "[\x00-\x1f\x60]" .Value) -}}{{fail (printf "%s: unsafe Go struct tag" .Location)}}{{- end -}}{{- end -}}
{{- if and .Repeated (eq .Type.Kind "enum") -}}{{fail (printf "%s: repeated enum needs a Go codec template change; not supported by the unchanged legacy template" .Location)}}{{- end -}}
{{- if or (not (matches "^[A-Za-z_][A-Za-z0-9_]*$" $name)) (has $fields $name) -}}{{fail (printf "%s: invalid or colliding Go field %s" .Location $name)}}{{- end -}}{{- set $fields $name true -}}
{{- range $keywords -}}{{- if eq . $name -}}{{fail (printf "Go field is a keyword: %s" $name)}}{{- end -}}{{- end -}}
{{- range list (concat $message.Name $name "Tag") (concat "NameOf" $message.Name $name) -}}{{- if has $symbols . -}}{{fail (printf "%s: duplicate generated symbol %s" $message.Location .)}}{{- end -}}{{- set $symbols . true -}}{{- end -}}
{{- end -}}{{- end -}}{{- end -}}
{{- range $keywords -}}{{- if has $symbols . -}}{{fail (printf "Go type is a keyword: %s" .)}}{{- end -}}{{- end -}}
{{end}}

{{define "fn/validateGoFilename"}}
{{- $name := index .Args 0 -}}{{- $location := index .Args 1 -}}
{{- if or (hasPrefix $name "_") (matches "_(test|aix|android|darwin|dragonfly|freebsd|hurd|illumos|ios|js|linux|nacl|netbsd|openbsd|plan9|solaris|wasip1|windows|zos|386|amd64|amd64p32|arm|arm64|loong64|mips|mipsle|mips64|mips64le|ppc64|ppc64le|riscv64|s390x|sparc64|wasm)$" $name) -}}{{fail (printf "%s: Go filename would be ignored or platform-specific: %s" $location $name)}}{{- end -}}
{{end}}
