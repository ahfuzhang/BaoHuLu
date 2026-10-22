# 架构迁移验证（2026-10-12）

## 状态与基线

**数据生成主线已迁移；整份计划尚未完成。** RPC 通用 handler 的必要正文变化仍待单独 review；旧 codec 的失败不混入本次机械拆分修复。

实施前工作区 HEAD 为 `5ddaa30`。旧工具单独构建，旧/新产物分别放入临时独立目录，不复用仓库 build/。使用 macOS arm64 / Apple M5、Go 1.27.0、.NET SDK 10.0.101；项目目标沿用 go 1.23 / net10.0。跨机器和 linux/amd64 未在本轮测量。

原仓库只有 Demo.proto，其中包含 AllTypes、ValueTypes、map/repeated/递归/枚举/大 tag/AsMap/AsArray/YAML/URL。`internal/integration/testdata/AllTypes.proto` 是它的固定副本，文件头新增三个生成开关；比较源码注释常量时旧工具也使用同一副本，避免行号差异。

## 已执行结果

| 验证 | 结果 |
| --- | --- |
| 核心/CLI/解析/模板生成测试 | 通过；含严格配置、路径冲突、人工改动保护、恢复、前向引用、统一 service/RPC 上下文 |
| 静态检查与并发检查 | go vet、CLI/core/parser 的 race、git diff --check 通过 |
| 16 种开关组合 | 两种语言各 8 种组合生成通过；project=on 的组合真实编译，含 test-only / bench-only |
| 资源来源 | 内置/外部输出逐字节等价；仓库外执行内置 Go、C#、RPC 包成功 |
| 最低 Go 工具链 | Go 1.23.12 仓库测试及 AllTypes 普通生成模块编译通过；不要求 jsonv2 |
| D16 模板区块原文 | 45 个区块 SHA-256 与固定基线一致，包含 Go、C#、RPC；测试记录原/新文件范围 |
| AllTypes Go 独立模块编译 | 通过；普通与 jsonv2 benchmark 文件可编译 |
| AllTypes Go 全部生成测试 | 229 pass / 11 fail，与旧版失败集合一致 |
| AllTypes C# 主/测试/benchmark 项目 | 均 build 成功；GrpcGen 项目成功 |
| AllTypes C# 全部生成测试 | 158 pass / 3 fail，与旧版失败集合一致 |
| Go 类型与函数 AST | 旧/新全部非测试类型、函数及方法正文一致（不仅是公开签名） |
| C# 类型成员与方法语法 token | 372 项一致；忽略文件边界、空白、注释和 using 排列 |
| C# 新旧 JSON/Protobuf 双向互验 | AllTypes 有界递归、单条 map entry、scalar/enum/bytes/repeated 向量通过 |
| Google.Protobuf 独立 oracle | 新编码→官方解析/重编码→新解码通过 |
| Go/C# 共享 AllTypes 向量 | 双向 JSON/Protobuf、Google.Protobuf 重编码、Clone/Reset 通过；无符号高区间的历史缺陷单列 |
| 生命周期 | Clone 后 Reset 原 reader 并覆盖原二进制，副本不变；原生成测试继续覆盖其他复用场景 |
| Shapes 小型 fixture | 前向引用、map/repeated、递归、空类型跳过、decimal；Go 编译与 C# build 通过 |
| RPC 兼容模板 | 统一引擎可生成；未宣称可编译或通用业务适配已完成 |

互验向量不包含 null map value 和显式空递归 Wrapper：旧实现通过 protobuf 往返会归一化为默认/缺省 message，不能要求它保留对象存在性。这不是此次迁移新增语义。

### 保留的历史失败

Go 的 11 个测试是 `TestChildFromJSONScalarTypeError`、`TestAllTypesFromJSONScalarTypeError`、`TestAllTypesFromJSONMsgTypeError`、`TestAllTypesFromJSONMapTypeError`、`TestAllTypesFromJSONArrayTypeError`、`TestValueTypesFromJSONScalarTypeError`、`TestHelloRequestFromJSONMsgTypeError`、`TestHelloResponseFromJSONScalarTypeError`、`TestHelloResponseFromJSONMsgTypeError`、`TestSeriesFromJSONMapTypeError`、`TestSeriesFromJSONArrayTypeError`。这些测试期望 null 报类型错误，而既有实现接受 null。

C# 的 3 个测试是 `ArrayTypeTests.FromJSON_InvalidInput_ReturnsError`、`AllTypesTests.YAMLRoundtrip_PreservesAllFields`、`AllTypesTests.YAMLRoundtrip_SpecialStrings_ThirdPartyValidated`。没有删除或跳过这些测试来制造全绿。

跨语言新增边界探测另发现旧 C# JSON 缺陷：`uint64` map key / `fixed64` 大于 Int64.MaxValue 时被转为有符号值，`uint32` / `fixed32` 高区间 JSON 解析按 int 而失败。新旧正文及行为一致，protobuf 解码仍正确。runner 显式验证并打印 KNOWN BASELINE LIMITATION，不把这些值列作跨语言 JSON 已支持。普通互通向量仍覆盖全部 scalar、64 位有符号最小值/uint64 最大值、所有 AllTypes map key、枚举、bytes、Unicode/转义、大 tag、有界递归。

Go 对 64 位整数仅在超出 JSON 安全整数区间时加引号，C# 一律加引号。跨语言比较只对五个已知 64 位整数字段规范化字符串/数字表示；同语言新旧比较保持严格，不放宽一般字符串字段。无符号上界缺陷不通过这种规范化掩盖。

## Go 同机性能

命令筛选 AllTypes/ValueTypes 的 JSON/Protobuf 编解码，`-benchmem -benchtime=1s -count=1`，旧/新交替各运行 5 次，表中为中位数 ns/op。首次 200ms 测量有两个项目超过 5%，复测后全部在建议阈值内；零分配场景仍为零。

| 场景 | 旧 ns/op | 新 ns/op | 变化 | 旧/新 allocs/op |
| --- | ---: | ---: | ---: | ---: |
| AllTypes ToJSON | 86505 | 84286 | -2.57% | 0 / 0 |
| AllTypes FromJSON | 231504 | 241167 | +4.17% | 62 / 62 |
| AllTypes ToProtobuf | 66743 | 66577 | -0.25% | 0 / 0 |
| AllTypes FromProtobuf | 89110 | 89187 | +0.09% | 66 / 66 |
| ValueTypes ToJSON | 85.59 | 85.14 | -0.53% | 0 / 0 |
| ValueTypes FromJSON | 269.3 | 274.9 | +2.08% | 0 / 0 |
| ValueTypes ToProtobuf | 17.43 | 17.87 | +2.52% | 0 / 0 |
| ValueTypes FromProtobuf | 31.10 | 31.57 | +1.51% | 0 / 0 |

Go 1.27 超过当前 Sonic 的优化支持范围，工具会退回 encoding/json；两边环境一致，以上选择的是 Generated 与 Protobuf 基准，不把 Sonic 性能用作迁移结论。C# BenchmarkDotNet 仅完成构建，不声称已有等价的运行性能统计。

## 生成成本（独立于生成代码运行性能）

同一 AllTypes 输入，test/benchmark/project 全开，旧/新交替各运行 5 次，每次写全新目录、包含进程启动和落盘。初次测量新 Go/C# 中位数为 1222/950 ms；发现按输出重复构模后，新增通用、显式的文件级 memo，并仅在新 partials 中复用只读模型，未修改旧模板区块。复测如下：

| 目标 | 旧中位数 ms | 新中位数 ms | 旧/新文件数 | 旧/新总字节数 |
| --- | ---: | ---: | ---: | ---: |
| Go | 61.52 | 590.16 | 13 / 115 | 527380 / 557719 |
| C# | 83.26 | 491.76 | 42 / 98 | 656228 / 690216 |

新耗时原始样本 Go 为 929.75/546.19/616.28/590.16/531.95 ms，C# 为 462.79/492.74/475.66/506.63/491.76 ms。文件数/字节数排除 .hulu-* 管理文件；新耗时包含严格 check、模板计算、更多文件及事务持久化，旧版无对应完整保护。**生成端仍显著慢于旧版，不属于“小幅增加”，需进一步 profiling/优化或明确接受；不能拿运行时 ≤5% 的结果掩盖这个差异。** 暂未测量两版产物的可比编译耗时/程序集体积。

缓存后额外一次 `/usr/bin/time -p`：Go real/user/sys 为 0.53/0.05/0.04 秒，C# 为 0.42/0.03/0.03 秒。墙钟时间远高于 CPU 时间，说明剩余时间不能全部归因于模板运算；不能为降低数字而取消校验、文件保护或同步写入。

## 复现

```sh
make test
make generated-build
cd build/generated/golang
GOEXPERIMENT=jsonv2 go test ./...
GOEXPERIMENT=jsonv2 go test -run '^$' -bench 'Benchmark(ToJSON|FromJSON|ToProtobuf|FromProtobuf)_(AllTypes|ValueTypes)(_Generated)?$' -benchmem -benchtime=1s -count=5
```

完整生成测试仍会报告上述历史失败。C# 在仓库根运行：

```sh
dotnet test build/generated/csharp/Tests/AllTypes.Tests.csproj --nologo
dotnet build build/generated/csharp/Benchmarks/AllTypes.Benchmarks.csproj --nologo
```

原基线工具/产物保存在本机 `/tmp/baohulu-migration.x42EWD/`（临时文件非永久依赖）。需要长期重现时从固定旧提交在独立目录构建，不回退当前工作区。

Go AST 对比：设置 `BAOHULU_BASELINE_GO` 与 `BAOHULU_MIGRATED_GO`，运行 `go test ./internal/integration -run TestGoBaselineDeclarations -v`。C# 互验：

```sh
dotnet run --project internal/integration/csharp-contract/Contract.csproj \
  -p:GeneratedDir=/absolute/new-csharp -p:BaselineDir=/absolute/old-csharp \
  -- /absolute/old-csharp-from-AllTypes-fixture /absolute/new-csharp
```

runner 需要新 AllTypes.dll、旧 Demo.dll、同版 QiWa.Common 和新 benchmark 的 GrpcGen.dll。语法对比的旧源码目录必须由同一 AllTypes.proto 生成。不同 assembly 用 extern alias 隔离。

共享向量（先运行 `make generated-build`，以下在仓库根执行）：

```sh
(cd internal/integration/go-contract && go run . write /tmp/hulu-vectors)
dotnet run --project internal/integration/csharp-contract/Contract.csproj \
  -p:GeneratedDir="$PWD/build/generated/csharp" -p:BaselineDir=/absolute/old-csharp \
  -- --vectors /tmp/hulu-vectors
(cd internal/integration/go-contract && go run . verify /tmp/hulu-vectors)
BAOHULU_BUILD_GENERATED=1 go test ./internal/integration -run 'TestCapabilityMatrix|TestBuiltinExternalEquivalence' -v
GOTOOLCHAIN=go1.23.12 go test ./cmd/... ./internal/... ./dependencies/... ./templates/...
```

## 仍需完成 / review

1. **RPC D16 例外**：修正不存在的 `req.Clone()`，移除 `rsp.Code/Message` 假设，业务改为独立人工 partial 回调，服务/方法符号消歧，再做真实 QiWa 服务端编译。当前只是兼容生成，不是验收完成。
2. **Go repeated enum**：旧正文按 message 处理导致不可编译；新模板先明确报错。`RepeatedEnum.proto` 固定该边界，不能把模型解析通过当作 codec 支持完成。
3. **Go 多 proto YAML**：旧私有注释常量按源行号命名，跨任务可能同名。需要修改常量定义及引用正文；尚未进行这项 D16 适配。
4. 生成耗时显著增加，虽已通过文件级缓存降低，仍需继续优化或审核接受。完整 C# 性能、linux/amd64/AOT 矩阵、编译耗时/生成程序集体积统计未完成；最低 Go 1.23.12 已验证。

以上问题没有藏进新 Go 语言适配器；暂停相关模板正文改变，已授权且独立的核心、数据模板、CLI、测试和文档继续完成。
