# BaoHuLu 架构改造待确认问题

本文根据 `doc/prompts/plan.md` 整理。问题分为“缺乏信息”“存在歧义”“缺乏例子”三类；建议先回答前两类，因为答案可能改变模板协议、输出布局和验收范围。可直接在每个“回答”后补充内容。

## 1. 缺乏的信息

### I-01 支持的工具链边界是什么？

计划尚未给出最低 Go 版本、最低 .NET SDK/Target Framework、必须支持的操作系统与 CPU 架构。请明确首版必须支持的版本和平台，以及哪些平台只要求生成代码、不要求在 CI 中构建。

回答：按照上一版的定义

### I-02 项目文件需要哪些元数据，数据从哪里取得？

Golang 的项目文件是 `go.mod`，C# 的项目文件是 `.csproj`。请明确 Go module path、Go 版本、C# `TargetFramework`、`RootNamespace`、`AssemblyName`、包版本和项目引用分别来自 proto option、模板固定值、输入文件名，还是其他来源。新版 CLI 已不提供这些参数，如果 proto 中也没有对应信息，需要确定默认规则。

回答：按照上一版的默认定义

### I-03 `RenderContext` 的完整稳定字段是什么？

计划只列出了 `Schema`、`File`、`Message`、`Enum`、`Service`、`RPC`、`TypeRef`、`Analysis`、`Options` 的顶层关系，但没有完整定义 `Message`、`Field`、`Enum`、`Analysis`、注释和 SourceLocation 的字段、类型及空值语义。请确认首版需要向外部模板承诺的最小字段集合，以及以后增加或删除字段时是否需要模板协议版本。

回答：空类型不生成代码

### I-04 external 类型需要保存哪些信息？

对于同包另一次生成提供的 external message/enum，请明确模型至少要保留哪些信息，模板应如何得到目标类型名、namespace/package、import 和 codec 方法名。也请明确哪些情况只告警，哪些情况必须阻止生成。

回答：

### I-05 所有 extension 的正式语法清单是什么？

除 `@gen_test_code`、`@gen_benchmark_code`、`@gen_project_file` 外，计划还涉及 `Deprecated`、`VarName`、`jsonName`、`yamlName`、`tag`、`decimal`、`AsMap`、`AsArray`、`yaml`、`UrlValues`、`path` 等。请给出首版真正支持的名称、作用域、取值格式、默认值、大小写规则、重复声明规则和冲突规则；`@formName`、`@method` 等仅有文档但尚未实现的名称是否保留也需要确认。

回答：

### I-06 `.hulu-manifest.json` 如何标识生成单元？

计划排除了仅用 proto basename 或机器绝对路径作为身份，但尚未确定最终算法。请明确输入文件、模板包、输出根和模板版本如何组成稳定 ID；外部模板目录移动后是否仍视为同一生成单元；两个 proto 共用输出根时谁拥有共享 support、`.gitignore`、`go.mod` 或 `.csproj`。

回答：

### I-07 RPC 首版依赖的 QiWa 契约是什么？

请明确首版服务端模板必须适配的 QiWa runtime/包版本、服务接口、路由注册方式、context/handler 签名、错误模型和项目引用。还需要列出哪些文件由生成器管理，哪些 partial/handler 文件由用户维护。

回答：

## 2. 模棱两可的地方

### A-01 “不改已有模板正文”与新模板上下文冲突时怎么办？

D16 要求已有 `.tpl` 只机械拆分，但统一 `RenderContext`、移除语言 helper 或修复作用域后可能无法直接执行旧正文。请确认此时是阻塞整个迁移，还是允许在单独 review 后做最小正文修改；如果允许，批准标准是什么。

回答：

### A-02 “按函数拆分”是否仍允许 From/To 模板包含多个函数？

计划目前把 `ToProtobuf`、`ToProtobufByAppend`、`ToProtobufVT`、`ProtobufSize` 等放入同一个 `ToProtobuf` 模板。请确认“以函数为单位拆分”是指每个函数一个 `.tpl`，还是指函数区块保持完整、多个相关函数仍可归入一个 From/To `.tpl`。

回答：

### A-03 项目文件默认 `on`，但模板包没有 `project` 映射时应如何处理？

当前计划写成：缺省 `@gen_project_file` 时，如果自定义模板包没有 `project` 映射则不报错；只有显式 `@gen_project_file=on` 才报“不支持”。请确认这是否是预期行为，RPC-only 模板包是否也遵循这一规则。

回答：

### A-04 项目文件关闭，但测试或 benchmark 打开时应生成什么？

当 `@gen_project_file=off` 且 `@gen_test_code=on` 或 `@gen_benchmark_code=on` 时，请确认仍生成测试/benchmark 源文件，只是不生成 `go.mod`/`.csproj`；还是应报组合冲突，避免产生无法独立构建的目录。

回答：

### A-05 create-only 项目文件过期时只告警是否足够？

`go.mod` 和 `.csproj` 被定义为 create-only。首次生成后再打开测试、benchmark 或 RPC，旧项目文件可能缺少新依赖。请确认生成器应始终只告警并要求人工修改，还是当文件内容仍等于上次清单时允许自动更新。

回答：

### A-06 生成之前的内部 check 校验到什么层级？

独立 `check` 命令已按要求移除。目前先校验 proto 语法、中立模型和 extension 值，再由所选模板包及通用 planner 检查能力、目标名称和输出路径；全部通过才写入。请确认这一内部阶段划分是否符合预期。

回答：

### A-07 deprecated 声明的生成规则是什么？

计划明确测试只针对“未废弃 message”，但没有完全说明 deprecated message、field、enum、service、method 是否仍生成类型和 codec，其他声明引用 deprecated 类型时是告警还是错误。请逐类确认。

回答：

### A-08 多个 proto 共用输出目录时，共享文件冲突如何解决？

计划提到“唯一所有者或显式合并策略”，但没有确定首版规则。请确认 `go.mod`、`.csproj`、测试 support、benchmark support 和 `.gitignore` 是首个生成单元拥有、固定主 proto 拥有、内容相同即可共享，还是禁止多个生成单元写同一路径。

回答：

### A-09 模板本次渲染为空时，旧文件是否应删除？

计划规定空白渲染不生成文件，但未明确上一次生成的 managed 文件在本次变为空时，是按 prune 删除、保留并告警，还是视为模板错误。请确认该生命周期规则。

回答：

### A-10 RPC 数据类型与 RPC 代码的命名不一致时何时报错？

计划不读取类型 manifest，也不扫描既有源码，最终依赖项目编译发现不一致。请确认生成阶段是否完全不做检查，还是内置 C# 数据模板与 QiWa.rpc 模板之间需要共享一个可校验的命名契约并在渲染前诊断。

回答：

## 3. 缺乏例子的地方

### E-01 请提供 extension 组合及预期输出目录例子

需要至少一个完整 proto，分别展示三个 extension 缺省、全部 `on`、项目 `off` 但测试/benchmark `on`，并列出 Golang 和 C# 各自应生成及不应生成的文件树。

回答或示例位置：

### E-02 请提供完整的模板包配置例子

现有 YAML 草案主要覆盖数据类型。请提供或确认一个包含 `message`、`enum`、codec、lifecycle、tests、benchmarks、project、service、rpc 的完整 `tpl_config.yaml`，以及一个只含最小 message 模板的第三方模板包例子。

回答或示例位置：

### E-03 请提供实际 `RenderContext` 数据例子

需要一个 message 渲染上下文和一个 service 渲染上下文。service 例子应包含多个 method、复用同一 request/response、一个 external 类型和多个 `@path`，并展示 `Methods`、`ReferencedTypes`、`RPC.Request/Response`、`TypeRef.Message` 的实际值。

回答或示例位置：

### E-04 请提供已有 `.tpl` 机械拆分的前后例子

需要用一个真实的小模板展示：原文件、拆分后的类型模板、函数模板、允许新增的 `define/end`、package/import 边界，以及哪些改动会被判定为“修改模板正文”而必须另行 review。

回答或示例位置：

### E-05 请提供 manifest 和文件生命周期例子

需要展示首次生成、无变化重跑、人工修改 managed 文件、打开/关闭测试、删除 message、模板渲染为空、两个 proto 共用输出根以及写入中途失败后的 `.hulu-manifest.json` 变化和预期诊断。

回答或示例位置：

### E-06 请提供名称转换与冲突例子

需要覆盖 proto 名称、`@VarName` 覆盖、Go/C# 关键字、大小写折叠、同名 service method、特殊 Go 文件后缀和 Windows 保留名，并给出期望生成名称或明确错误。

回答或示例位置：

### E-07 请提供 external 类型和 RPC 分次生成例子

需要两个 proto 或两次生成命令，展示数据类型先生成、RPC 后生成时，external request/response 如何命名、需要哪些 namespace/import/project 引用，以及命名不匹配时用户能看到什么错误。

回答或示例位置：

## 4. 实施中必须单独 review 的正文改动

### R-01 是否批准 RPC 的最小 D16 例外？

旧 MethodContext 调用不存在的 `req.Clone()`，应改为 `Clone(ref dst)`；旧服务实现假定响应具有 Code/Message，无法适用于任意 response。建议把业务处理放入独立人工 partial handler，生成模板仅处理路由/编解码，并给 context 符号加服务前缀。是否批准这组与机械拆分分开的正文适配？未批准前保持原区块哈希，不宣称 RPC 已可编译。

回答：

### R-02 Go repeated enum 与跨 proto YAML 常量是否另开修复？

旧 repeated enum codec 按 message 处理，生成物不能编译；目前提前明确报错。旧 YAML 私有常量只按源行号命名，多 proto 共用包可能碰撞。两项均需改模板正文，建议单独修复与审核，不混入机械拆分。

回答：

### R-03 历史 codec 失败是否保持独立处理？

AllTypes 新旧 Go 同为 11 项失败、C# 同为 3 项失败；新增跨语言边界也确认 C# 高区间无符号 JSON 存在旧缺陷。详见 GenerationValidation.md。当前未修改这些旧正文，是否另列后续修复而不改变本次 D16 约束？

回答：

### R-04 是否接受首次完整生成的成本增加？

同机 AllTypes、三个开关全开时，文件级缓存优化后新 Go/C# 首次生成中位数约 590/492 ms，旧版为 62/83 ms；新流程包括更多拆分文件、强制 check 与事务持久化。生成代码运行性能未出现超过 5% 的变化，不能混同这两个指标。是否接受当前生成耗时，还是必须在保留写入保护的前提下继续优化后才能验收？

回答：
