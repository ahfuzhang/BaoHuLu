/*
client 设计思路:
* 第一版先针对 http2
* 每个 http client 只有 1 个 tcp 连接
* 每个 service 对应着一组 http client (而不是每个 method 对应着一组)
* 每个 http client 初始化的时候，传入 ip:port
  - 便于在云的环境中，针对域名，解析出一堆 ip:port
* 每个 service 一个域名
  - 每个域名一个 map
    - key 是 ip:port
    - value 是 HttpClient 对象的数组
  - 存在一个 client 选择器的对象
    - 以一致性 hash 算法，根据 key，选择 ip:port （或者更丰富的路由选择算法）
    - 根据 HttpClient 数组，以 rand robin 来选择同一个 ip:port 上的 tcp 的其中一条
      - 当某个 http2 的 tcp 的 stream count 超过一定数量，进行扩容（HttpClient 自带的能力）
  - dns 更新策略
* 可观测性：
* 重试，容灾调度

*/
#pragma warning disable CS1591 // Missing XML comment for publicly visible type or member

namespace Generated.Demo.Http1;

using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Reflection.Metadata;
using System.Threading;
using System.Threading.Tasks;
using QiWa.Common;
using QiWa.Compress;

enum CompressType
{
    NotCompress = 0,
    Zstd = 1,
    Gzip = 2
}

enum EncodeType
{
    Unknown = 0,
    JSON = 1,
    Protobuf = 2
}

// 先支持 http2
public class LoginClient
{
    private RentedBuffer buf = new RentedBuffer(1024);
    //public string Addr;  // ip:port
    private readonly System.Net.IPEndPoint Addr;
    private readonly string Host;
    private readonly string Path;
    readonly Uri uri;
    internal readonly SocketsHttpHandler handler;
    private HttpClient client;

    private LoginClient()
    {

    }

    public LoginClient(System.Net.IPEndPoint addr, string host, string path)
    {
        this.Addr = addr;
        this.Host = host;
        this.Path = path;
        this.uri = new Uri($"http://{host}{path}");
        handler = new SocketsHttpHandler
        {
            MaxConnectionsPerServer = 1,
            EnableMultipleHttp2Connections = false,
            PooledConnectionIdleTimeout = TimeSpan.FromMinutes(5),
            KeepAlivePingDelay = TimeSpan.FromSeconds(60),
            KeepAlivePingPolicy = HttpKeepAlivePingPolicy.Always,
            AutomaticDecompression = DecompressionMethods.None,

            ConnectCallback = async (context, cancellationToken) =>
            {
                var socket = new Socket(SocketType.Stream, ProtocolType.Tcp)
                {
                    NoDelay = true
                };
                try
                {
                    await socket.ConnectAsync(
                        //new IPEndPoint(IPAddress.Parse("1.2.3.4"), 443),
                        this.Addr,
                        cancellationToken);
                    return new NetworkStream(socket, ownsSocket: true);
                }
                catch
                {
                    socket.Dispose();
                    throw;
                }
            }
        };
        //
        client = new HttpClient(handler)
        {
            DefaultRequestVersion = HttpVersion.Version20,
            DefaultVersionPolicy = HttpVersionPolicy.RequestVersionExact
        };
    }

    //internal readonly SocketsHttpHandler 

    //private HttpClient 
    //public LoginRequest Request;
    //public ReadonlyLoginResponse Response;


    public Error Login(ref readonly LoginRequest req, ref ReadonlyLoginResponse rsp, 
        CancellationToken ct, CompressType cpmpress=CompressType.Zstd, EncodeType encodeType=EncodeType.Protobuf)
    {
        if (encodeType==EncodeType.Unknown)
        {
            return Error.WithLoc(1, "encode type must json/protobuf");
        }
        try
        {
            var req = new HttpRequestMessage(HttpMethod.Post, this.uri)
            {
                Version = client.DefaultRequestVersion,
                VersionPolicy = client.DefaultVersionPolicy,
            };
            req.Headers.TryAddWithoutValidation("Accept-Encoding", "zstd,gzip");
            // Serialize request
            buf.Length = 0;
            byte[] rawBody;
            if (encodeType==EncodeType.Protobuf)
            {
                req.ToProtobuf(ref buf);
                rawBody = buf.AsSpan().ToArray();
            }
            else
            {
                req.ToJSON(ref buf);
                rawBody = buf.AsSpan().ToArray();
            }
            Interlocked.Add(ref s_totalBytesSent, rawBody.Length);

            // Compress if requested
            byte[] sendBody = rawBody;
            if (!string.IsNullOrEmpty(s_compress))
            {
                if (s_compress == "gzip")
                {
                    var (cbuf, gzipCompressErr) = GzipCompressor.Compress(rawBody);
                    if (gzipCompressErr.Err())
                    {
                        Console.WriteLine($"[ERROR] GzipCompressor.Compress failed: {gzipCompressErr}");
                        Interlocked.Increment(ref s_totalErrors);
                        continue;
                    }
                    gzipBuf.Dispose();
                    gzipBuf = cbuf;
                    sendBody = gzipBuf.AsSpan().ToArray();
                }
                else
                {
                    zstdBuf.Length = 0;
                    var zstdCompressErr = ZstdCompressor.Compress(ref zstdBuf, rawBody);
                    if (zstdCompressErr.Err())
                    {
                        Console.WriteLine($"[ERROR] ZstdCompressor.Compress failed: {zstdCompressErr}");
                        Interlocked.Increment(ref s_totalErrors);
                        continue;
                    }
                    sendBody = zstdBuf.AsSpan().ToArray();
                }
                Interlocked.Add(ref s_totalCompressedBytesSent, sendBody.Length);
            }

            // Build request
            using var content = new ByteArrayContent(sendBody);
            content.Headers.ContentType = s_encodeType == "protobuf" ? s_contentTypeProtobuf : s_contentTypeJson;
            if (!string.IsNullOrEmpty(s_compress))
                content.Headers.ContentEncoding.Add(s_compress);

            req.Content = content;

            // Send and measure latency
            long t0 = Stopwatch.GetTimestamp();
            using var resp = await client.SendAsync(req, HttpCompletionOption.ResponseHeadersRead, ct);

            if (!resp.IsSuccessStatusCode)
            {
                Console.WriteLine($"[ERROR] HTTP {(int)resp.StatusCode} {resp.StatusCode} from {addr}");
                Interlocked.Increment(ref s_totalErrors);
                continue;
            }

            // Read response body into reusable buffer (zero allocation per iteration)
            respBuf.Length = 0;
            using (var respStream = await resp.Content.ReadAsStreamAsync(ct))
            {
                long? cl = resp.Content.Headers.ContentLength;
                if (cl.HasValue)
                {
                    int len = (int)cl.Value;
                    if (respBuf.Data.Length < len)
                        respBuf.Extend(len - respBuf.Data.Length);
                    int offset = 0, remaining = len;
                    while (remaining > 0)
                    {
                        int n = await respStream.ReadAsync(respBuf.Data.AsMemory(offset, remaining), ct);
                        if (n == 0) break;
                        offset += n;
                        remaining -= n;
                    }
                    respBuf.Length = offset;
                }
                else
                {
                    // chunked transfer: stream until EOF, grow buffer as needed
                    int offset = 0;
                    while (true)
                    {
                        int free = respBuf.Data.Length - offset;
                        if (free == 0)
                        {
                            respBuf.Extend(respBuf.Data.Length);
                            free = respBuf.Data.Length - offset;
                        }
                        int n = await respStream.ReadAsync(respBuf.Data.AsMemory(offset, free), ct);
                        if (n == 0) break;
                        offset += n;
                    }
                    respBuf.Length = offset;
                }
            }

            long latencyUs = (Stopwatch.GetTimestamp() - t0) * 1_000_000L / Stopwatch.Frequency;

            Interlocked.Increment(ref s_latencyBuckets[GetBucket(latencyUs)]);
            Interlocked.Add(ref s_totalBytesReceived, respBuf.Length);

            // Decompress response if needed
            ReadOnlySpan<byte> decodeSpan = respBuf.Data.AsSpan(0, respBuf.Length);
            if (!string.IsNullOrEmpty(s_compress) && resp.Content.Headers.ContentEncoding.Count > 0)
            {
                string enc = resp.Content.Headers.ContentEncoding.First();
                if (enc == "gzip")
                {
                    var (dbuf, gzipDecompressErr) = GzipCompressor.Uncompress(respBuf.Data.AsSpan(0, respBuf.Length));
                    if (gzipDecompressErr.Err())
                    {
                        Console.WriteLine($"[ERROR] GzipCompressor.Uncompress failed: {gzipDecompressErr}");
                        Interlocked.Increment(ref s_totalErrors);
                        continue;
                    }
                    gzipBuf.Dispose();
                    gzipBuf = dbuf;
                    decodeSpan = gzipBuf.AsSpan();
                }
                else if (enc == "zstd")
                {
                    zstdBuf.Length = 0;
                    var zstdDecompressErr = ZstdCompressor.Uncompress(ref zstdBuf, respBuf.Data.AsSpan(0, respBuf.Length));
                    if (zstdDecompressErr.Err())
                    {
                        Console.WriteLine($"[ERROR] ZstdCompressor.Uncompress failed: {zstdDecompressErr}");
                        Interlocked.Increment(ref s_totalErrors);
                        continue;
                    }
                    decodeSpan = zstdBuf.AsSpan();
                }
                Interlocked.Add(ref s_totalDecompressedBytesReceived, decodeSpan.Length);
            }

            // Deserialize response (reuse state.Response via Reset)
            state.Response.Reset();
            if (s_encodeType == "protobuf")
            {
                var protoRespErr = state.Response.FromProtobuf(decodeSpan);
                if (protoRespErr.Err())
                {
                    Console.WriteLine($"[ERROR] Response.FromProtobuf failed: {protoRespErr}");
                    Interlocked.Increment(ref s_totalErrors);
                    continue;
                }
            }
            else
            {
                var jsonRespErr = state.Response.FromJSON(decodeSpan);
                if (jsonRespErr.Err())
                {
                    Console.WriteLine($"[ERROR] Response.FromJSON failed: {jsonRespErr}");
                    Interlocked.Increment(ref s_totalErrors);
                    continue;
                }
            }

            Interlocked.Increment(ref s_totalRequests);
        }
        catch (OperationCanceledException) { break; }
        catch (Exception ex) { Console.WriteLine($"[ERROR] Unexpected exception: {ex}"); Interlocked.Increment(ref s_totalErrors); }
        return default;
    }
}


