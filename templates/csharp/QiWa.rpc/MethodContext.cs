#pragma warning disable CS1591 // Missing XML comment for publicly visible type or member

namespace Generated.Demo;

using QiWa.Common;
using QiWa.KestrelWrap;

class HelloContext : ContextBase, QiWa.Common.IResettable
{
    public ReadonlyHelloRequest Request;
    public HelloResponse Response;
    //todo: 在这里定义业务处理中的局部变量，从而最终做到 0 alloc

    public void Reset()
    {
        base.Reset();
        Request.Reset();
        Response.Reset();
        //todo: 局部变量的 reset 写在这里
    }

    public async ValueTask<Error> Run()
    {
        // todo: 业务代码写在这里
        // ref readonly ReadonlyHelloRequest req = ref Request;
        // req.Reset();
        // ref HelloResponse rsp = ref Response;
        // rsp.Reset();
        Request.Reset();
        Response.Reset();
        return default;
    }
}
