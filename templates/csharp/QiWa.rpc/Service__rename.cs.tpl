#pragma warning disable CS1591 // Missing XML comment for publicly visible type or member

namespace {{.CsharpNamespace}};

using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.ObjectPool;
using QiWa.Common;
using QiWa.ConsoleLogger;
using QiWa.KestrelWrap;

public class {{.ServiceName}}Counters : QiWa.Common.IResettable
{
{{- range .Methods}}
    public UInt64 {{.MethodName}}RequestTotal;
    public UInt64 {{.MethodName}}DecodeErrorsTotal;
    public UInt64 {{.MethodName}}ExceptionsTotal;
    public UInt64 {{.MethodName}}LogicErrorsTotal;
{{- end}}

    public void Reset()
    {
{{- range .Methods}}
        {{.MethodName}}RequestTotal = 0;
        {{.MethodName}}DecodeErrorsTotal = 0;
        {{.MethodName}}ExceptionsTotal = 0;
        {{.MethodName}}LogicErrorsTotal = 0;
{{- end}}
    }
}

public class {{.ServiceName}}  // 这里是 service 的名字
{
{{- range .Methods}}
    internal static readonly DefaultObjectPool<{{.MethodName}}Context> {{.MethodName}}ContextPool = new DefaultObjectPool<{{.MethodName}}Context>(
        new ContextObjectPolicy<{{.MethodName}}Context>(),
        maximumRetained: ServerConfig.MaxCocurrentCount
    );
{{- end}}

    // ThreadLocal
    internal static readonly ThreadLocal<{{.ServiceName}}Counters> _threadLocal =
        new ThreadLocal<{{.ServiceName}}Counters>(() => new {{.ServiceName}}Counters(), trackAllValues: true);
    public static {{.ServiceName}}Counters Counters => _threadLocal.Value!;

    public static {{.ServiceName}}Counters SumCounters({{.ServiceName}}Counters? dst)
    {
        if (dst == null)
        {
            dst = new {{.ServiceName}}Counters();
        }
        foreach ({{.ServiceName}}Counters c in _threadLocal.Values)
        {
{{- range .Methods}}
            dst.{{.MethodName}}RequestTotal = Interlocked.Read(ref c.{{.MethodName}}RequestTotal);
            dst.{{.MethodName}}DecodeErrorsTotal = Interlocked.Read(ref c.{{.MethodName}}DecodeErrorsTotal);
            dst.{{.MethodName}}ExceptionsTotal = Interlocked.Read(ref c.{{.MethodName}}ExceptionsTotal);
            dst.{{.MethodName}}LogicErrorsTotal = Interlocked.Read(ref c.{{.MethodName}}LogicErrorsTotal);
{{- end}}            
        }
        return dst;
    }

    public static async Task HandleAsync(HttpContext context)
    {
        Interlocked.Increment(ref ContextBase.Counters.HttpRequestTotal);
        Error err = ContextBase.Validate(context);
        if (err.Err())
        {
            // 打日志
            ThreadLocalLogger.Current.Warn(
                Field.String("path"u8, context.Request.Path.Value ?? ""),
                Field.String("method"u8, context.Request.Method),
                Field.String("protocol"u8, context.Request.Protocol),
                Field.String(
                    (context.Request.HttpContext.Connection.RemoteIpAddress?.AddressFamily == System.Net.Sockets.AddressFamily.InterNetworkV6)
                        ? "client_ipv6"u8 : "client_ipv4"u8,
                    context.Request.HttpContext.Connection.RemoteIpAddress?.ToString() ?? ""),
                Field.Int64("error_code"u8, err.Code),
                Field.String("message"u8, err.Message)
            );
            // metrics 上报
            Interlocked.Increment(ref ContextBase.Counters.HttpBadRequestTotal);
            return;
        }
        // 判断请求路径
        byte[]? responseBytes;
        switch (context.Request.Path)
        {
{{- range .Methods}}
            case "/{{$.ServiceName}}/{{.MethodName}}":
{{- if .Path}}
            case "{{.Path}}":
{{- end}}
                {
                    Interlocked.Increment(ref Counters.{{.MethodName}}RequestTotal);
                    {{.MethodName}}Context ctx = {{.MethodName}}ContextPool.Get();
                    using var _ = new QiWa.Helper.ScopeGuard(() =>
                    {
                        {{.MethodName}}ContextPool.Return(ctx);
                        //todo: 上报处理时间
                    });
                    err = ctx.InitFromHttp(context);
                    if (err.Err())
                    {
                        // 打日志
                        ThreadLocalLogger.Current.Warn(
                            Field.String("path"u8, context.Request.Path.Value ?? ""),
                            Field.String("method"u8, context.Request.Method),
                            Field.String("protocol"u8, context.Request.Protocol),
                            Field.String(
                                (context.Request.HttpContext.Connection.RemoteIpAddress?.AddressFamily == System.Net.Sockets.AddressFamily.InterNetworkV6)
                                    ? "client_ipv6"u8 : "client_ipv4"u8,
                                context.Request.HttpContext.Connection.RemoteIpAddress?.ToString() ?? ""),
                            Field.Int64("error_code"u8, err.Code),
                            Field.String("message"u8, err.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref ContextBase.Counters.InitErrorsTotal);
                        return;
                    }
                    byte[]? reqRequest;
                    (reqRequest, err) = await ctx.ReadRequest().ConfigureAwait(true);
                    if (err.Err())
                    {
                        // 打日志
                        ctx.L!.Warn(
                            Field.Int64("error_code"u8, err.Code),
                            Field.String("message"u8, err.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref ContextBase.Counters.InitErrorsTotal);
                        return;
                    }
                    // 解码
                    err = ctx.Decode<Readonly{{.RequestType}}>(reqRequest!, ref ctx.Request);
                    if (err.Err())
                    {
                        // 打日志
                        ctx.L!.Warn(
                            Field.Int64("error_code"u8, err.Code),
                            Field.String("message"u8, err.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref Counters.{{.MethodName}}DecodeErrorsTotal);
                        return;
                    }
                    // 调用业务
                    try
                    {
                        // 加上计时
                        err = await ctx.Run().ConfigureAwait(true);
                    }
                    catch (Exception ex)
                    {
                        // 打日志
                        ctx.L!.Warn(
                            Field.Int64("error_code"u8, 65535),
                            Field.String("message"u8, ex.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref Counters.{{.MethodName}}ExceptionsTotal);
                        context.Response.StatusCode = 500;
                        return;
                    }
                    if (err.Err())
                    {
                        // 打日志
                        ctx.L!.Warn(
                            Field.Int64("error_code"u8, err.Code),
                            Field.String("message"u8, err.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref Counters.{{.MethodName}}LogicErrorsTotal);
                        return;
                    }
                    // 响应
                    (responseBytes, err) = ctx.Encode<{{.ResponseType}}>(ref ctx.Response);
                    if (err.Err())
                    {
                        // 打日志
                        ctx.L!.Warn(
                            Field.Int64("error_code"u8, err.Code),
                            Field.String("message"u8, err.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref ContextBase.Counters.EncodeErrorsTotal);
                        return;
                    }
                }
                break;
{{- end}}
            default:
                // 多个 service 如何处理?
                context.Response.StatusCode = 404;
                // 打日志  => 避免因为扫描路径而产生大量日志。此处避免输出太多日志
                // 数据上报  => 可以考虑使用 thread local
                Interlocked.Increment(ref ContextBase.Counters.HttpNotFoundErrorsTotal);
                return;
        }
        // 输出
        context.Response.StatusCode = 200;
        try
        {
            // ??? 网络发送的时间，是否需要记录
            await context.Response.Body.WriteAsync(responseBytes, context.RequestAborted).ConfigureAwait(false);
        }
        catch (OperationCanceledException ex)
        {
            ThreadLocalLogger.Current.Warn(
                Field.String("path"u8, context.Request.Path.Value ?? ""),
                Field.String("method"u8, context.Request.Method),
                Field.String("protocol"u8, context.Request.Protocol),
                Field.String(
                    (context.Request.HttpContext.Connection.RemoteIpAddress?.AddressFamily == System.Net.Sockets.AddressFamily.InterNetworkV6)
                        ? "client_ipv6"u8 : "client_ipv4"u8,
                    context.Request.HttpContext.Connection.RemoteIpAddress?.ToString() ?? ""),
                Field.Int64("error_code"u8, 65535),
                Field.String("message"u8, ex.Message),
                Field.String("exception"u8, "OperationCanceledException")
            );
            Interlocked.Increment(ref ContextBase.Counters.SendErrorsTotal);
            return;
        }
        // todo: 拦截器调用
    }
}
