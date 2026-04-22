#pragma warning disable CS1591 // Missing XML comment for publicly visible type or member

namespace {{.CsharpNamespace}};

using QiWa.Common;
using QiWa.KestrelWrap;

class {{.MethodName}}Context : ContextBase, QiWa.Common.IResettable
{
    public Readonly{{.RequestType}} Request;
    public {{.ResponseType}} Response;
    //todo: 在这里定义业务处理中的局部变量，从而最终做到 0 alloc

    public new void Reset()
    {
        base.Reset();
        Request.Reset();
        Response.Reset();
        //todo: 局部变量的 reset 写在这里
    }

    public async ValueTask<Error> Run()
    {
        // todo: 业务代码写在这里
        // ref readonly Readonly{{.RequestType}} req = ref Request;
        // req.Reset();
        // ref {{.ResponseType}} rsp = ref Response;
        // rsp.Reset();
        Request.Reset();
        Response.Reset();
        return default;
    }
}
