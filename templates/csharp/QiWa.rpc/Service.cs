#pragma warning disable CS1591 // Missing XML comment for publicly visible type or member

namespace Generated.Demo;

using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.ObjectPool;
using QiWa.Common;
using QiWa.ConsoleLogger;
using QiWa.KestrelWrap;

public class DemoServerCounters
{
    // todo: method
    public UInt64 HelloRequestTotal;
    public UInt64 HelloDecodeErrorsTotal;
    public UInt64 HelloExceptionsTotal;
    public UInt64 HelloLogicErrorsTotal;
    // todo: end of  method
}

public class DemoServer  // 这里是 service 的名字
{
    internal static readonly DefaultObjectPool<HelloContext> HelloContextPool = new DefaultObjectPool<HelloContext>(
        new ContextObjectPolicy<HelloContext>(),
        maximumRetained: ServerConfig.MaxCocurrentCount
    );

    // ThreadLocal
    internal static readonly ThreadLocal<DemoServerCounters> _threadLocal =
        new ThreadLocal<DemoServerCounters>(() => new DemoServerCounters(), trackAllValues: true);
    public static DemoServerCounters Counters => _threadLocal.Value!;

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
            // todo: 遍历 service 下的每个 method，在此处生成代码
            case "/service/Hello":  // 这里是每个 method 的路径
                {
                    Interlocked.Increment(ref Counters.HelloRequestTotal);
                    HelloContext ctx = HelloContextPool.Get();
                    using var _ = new QiWa.Helper.ScopeGuard(() =>
                    {
                        HelloContextPool.Return(ctx);
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
                    err = ctx.Decode<ReadonlyHelloRequest>(reqRequest!, ref ctx.Request);
                    if (err.Err())
                    {
                        // 打日志
                        ctx.L!.Warn(
                            Field.Int64("error_code"u8, err.Code),
                            Field.String("message"u8, err.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref Counters.HelloDecodeErrorsTotal);
                        return;
                    }
                    // 调用业务
                    try
                    {
                        // 加上计时
                        err = await ctx.Run().ConfigureAwait(true);  // todo: 这里要加异常处理
                    }
                    catch (Exception ex)
                    {
                        // 打日志
                        ctx.L!.Warn(
                            Field.Int64("error_code"u8, 65535),
                            Field.String("message"u8, ex.Message)
                        );
                        // 数据上报
                        Interlocked.Increment(ref Counters.HelloExceptionsTotal);
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
                        Interlocked.Increment(ref Counters.HelloLogicErrorsTotal);
                        return;
                    }
                    // 响应
                    (responseBytes, err) = ctx.Encode<HelloResponse>(ref ctx.Response);
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
            // todo: 遍历 service 的每个 method. 此处结束其中一个 method 的代码生成    
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
