#pragma warning disable CS1591 // Missing XML comment for publicly visible type or member

namespace Generated.Demo.Http1;

using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Threading;
using System.Threading.Tasks;
using QiWa.Common;
using QiWa.Compress;



public class LoginClient
{
    internal static SocketsHttpHandler handler = new SocketsHttpHandler
        {
            MaxConnectionsPerServer = 1,
            EnableMultipleHttp2Connections = false,
            PooledConnectionIdleTimeout = TimeSpan.FromMinutes(5),
            KeepAlivePingDelay = TimeSpan.FromSeconds(60),
            KeepAlivePingPolicy = HttpKeepAlivePingPolicy.Always,
            AutomaticDecompression = DecompressionMethods.None,
        };    

    private HttpClient client = new HttpClient(handler);
    //public LoginRequest Request;
    //public ReadonlyLoginResponse Response;
    private RentedBuffer buf = new RentedBuffer(1024);

    public Error Login(ref readonly LoginRequest req, ref ReadonlyLoginResponse rsp, CancellationToken ct)
    {
        return default;
    }
}


