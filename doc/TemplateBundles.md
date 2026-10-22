# 模板包协议 v1

## 流水线与目录

`cmd/hulu` 仅解析通用参数；`protofile.Parse` 解析一次并进行内部 check；`schema` 解析前向引用与分析依赖；`codegen` 加载配置、计划输出、完整渲染、验证所有权再提交。`templates/embed.go` 只提供资源。

没有 Go/C# 生成器包、语言 View struct、语言 switch、项目源码字符串或跨任务类型 manifest。文件级开关和语言无关协议事实保存在 Schema；目标命名、类型、Wrapper、literal、结构布局维度等由模板计算。

模板包必须有 `tpl_config.yaml`，只允许以下键：

```yaml
message: types/message.txt.tpl
enum: types/enum.txt.tpl
json:
  FromJSON: methods/from.txt.tpl
  ToJSON: methods/to.txt.tpl
protobuf:
  FromProtobuf: methods/from-proto.txt.tpl
  ToProtobuf: methods/to-proto.txt.tpl
yaml:
  FromYAML: methods/from-yaml.txt.tpl
  ToYAML: methods/to-yaml.txt.tpl
url_values:
  FromURLValues: methods/from-url.txt.tpl
  ToURLValues: methods/to-url.txt.tpl
lifecycle:
  Reset: methods/reset.txt.tpl
  Clone: methods/clone.txt.tpl
tests:
  "{message}.test.txt": tests/message.txt.tpl
benchmarks:
  "{file}.bench-support.txt": benchmarks/support.txt.tpl
project:
  "{file}.project": project/project.tpl
service: service/server.txt.tpl
rpc: method/context.txt.tpl
```

各部分可省略；值不能为空或 null。拒绝未知键、重复 YAML 键、多 YAML 文档、未知方法键、缺文件、重复非空 define 和越界路径。加载显式引用的文件及 `partials/` 下全部 `.tpl`；不会自动输出整个目录。包内不允许 symlink。

类型输出是 `原始类型名 + 模板扩展名`；方法是 `类型名.方法名 + 扩展名`。service/RPC 文件名见计划 9.2。tests/benchmarks/project 的键是相对输出路径，只替换 `{file}`、`{message}`；前者是不含扩展名的输入文件 basename。

enum 仅运行 enum 模板；未废弃且有有效字段的 message 才运行 message/codec 模板。YAML 传播到所有本地 message 引用；URL Values 不传播 map 值。空类型不生成，被其他字段/RPC 引用时诊断而不是输出缺失类型。

输出排序确定，跨平台大小写冲突、Windows 保留名、路径逃逸、文件/目录冲突在写入前报错。模板只渲染空白时不输出，该任务之前拥有的同名 managed 文件进入 prune。

## RenderContext

所有入口都接收同一种 `internal/codegen.RenderContext`：

| 字段 | 内容 |
| --- | --- |
| Schema、File | 完整中立模型；原始源码、package、options、imports、开关、注解 |
| Message / Enum | 对应类型入口唯一非空的类型指针；共享支持/项目入口均为空 |
| Service / RPC | service 入口只有 Service；rpc 入口两者都有，不伪装为 Message |
| TypeRef | 由 withTypeRef 显式设置的当前类型引用 |
| Analysis | 有序直接/值依赖、依赖顺序、YAML/URL 闭包 |
| Options | Source、TemplateDir、OutputDir、Version；不是语言专属配置 |
| Args | 模板函数的有序参数，普通入口为空 |

具体字段以 `internal/schema/schema.go` 为准。原始字段名、原始引用名称、完整符号 ID、源码位置、注释及注解均保留；不包含 GoType/MapKeyCS 等结果。

`File.OptionValues` 保存文件级 option 值的原始字节区间（Start 含、End 不含）；配合原始 Source 和标准 `slice` 可只替换目标 option，不误改同值的其他 option 或注释。C# benchmark 的 namespace 副本使用此中立事实，核心不拼接目标语言源码。

TypeRef.Kind 是 scalar/message/enum/external。本地引用的 Message/Enum 指向 Schema 中的唯一对象；external 只有原始名称及标识，生成前警告，不做 import 文件解析或符号扫描。Schema 在渲染期间只读，模板只可修改自身创建的 dict/list。

Service.Methods 保持声明顺序；ReferencedTypes 按方法顺序 request→response 去重，只收集直接引用，不替代 Schema 的完整类型表。

```gotemplate
{{range .Service.Methods}}
  {{render "rpc.signature" (withRPC $ .)}}
{{end}}
{{define "rpc.signature"}}
  {{.Service.Name}}.{{.RPC.Name}}:
  {{render "type.name" (withTypeRef . .RPC.Request)}}
{{end}}
{{define "type.name"}}{{.TypeRef.RawName}}{{end}}
```

`withRPC` 返回浅拷贝，设置 Service/RPC 并清空 Message/Enum/TypeRef；`withTypeRef` 只替换 TypeRef。`render` 仍传统一上下文，不计算语言名称。不存在 types.manifest。

## 模板函数与原正文兼容

除 Go text/template 内置函数外，提供 render、withRPC、withTypeRef、dict/list/get/has/set/append、字符串和整数运算、reverse/sortBy、matches、reachable、tagBytes、layout 等无语言规则的工具；完整签名见 `internal/codegen/functions.go`。

`layout` 接收模板指定的 Size/Align/Ptrdata/Number 整数描述，稳定排序并返回 Fields/Size/Ptrdata；核心不知道任何目标类型的尺寸。内置包保留旧版的 64 位布局估算。

为了不改写旧方法正文，包可定义 `fn/名称`。引擎把它绑定为同名函数，输入仍是 RenderContext，实参放在 Args。纯文本输出返回字符串；调用 `return` 可返回 bool/list/dict 等类型值，不经过 JSON 中间文件。

```gotemplate
{{define "fn/enabled"}}{{return false}}{{end}}
{{define "fn/label"}}{{.Message.Name}}/{{index .Args 0}}{{end}}
{{if enabled}}enabled{{end}}
{{label "writer"}}
```

`return` 只允许在模板函数中调用一次；它记录结果但不提前终止执行，不能同时输出非空白文本。函数/递归 render 限制 64 层，单次输出限制 64 MiB，未知字段使用 missingkey=error。无 shell、网络或任意文件读写函数。普通 named template 的 `template` 调用沿用标准库语义。

`memo "函数名"` 是显式选择的文件级计算缓存，只能调用包内 fn/*。计算时只传 Schema/File/Analysis/Options，其余选择器及 Args 为空；同次 Render、相同文件/Schema/Options 复用结果，下一次 Render 清空缓存。结果须视为只读，调用者不能 set/append 修改共享对象；每个输出如需可变属性，应另建 dict。缓存递归、覆盖内置函数名均报错。内置包仅用它缓存已完成的文件级模型，消息选择仍在各自上下文进行，不隐式缓存任意 helper。

内置 `model`/`csModel` 函数返回的旧属性字典仅存在于模板内；字段名沿用旧正文是兼容包装，不是第二套 Go 数据模型。RPC 的旧 `$` 根通过 named template 重绑定。

## 最小外部模板包

创建 `tpl_config.yaml` 内容 `message: message.txt.tpl`，以及 `message.txt.tpl` 内容：

```gotemplate
{{.File.Package}}.{{.Message.Name}}: {{len .Message.ActiveFields}} fields
```

执行 `hulu -src=input.proto -template.dir=该目录 -out=新目录` 即可；没有必选语言标识，也不需要 project 映射。输入没有显式启用该包不提供的能力时，文件级默认值可以直接使用。
