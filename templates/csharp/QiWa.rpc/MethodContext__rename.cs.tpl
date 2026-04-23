#pragma warning disable CS1591 // Missing XML comment for publicly visible type or member

namespace {{.CsharpNamespace}};

using System.Text;
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
        ref readonly var req = ref Request;
        ref var rsp = ref Response;
        // todo: write your bussiness logic code here
        rsp.Code = 0;
        var buf = new RentedBuffer(1024);
        string j;
        try
        {
            {{.RequestType}} temp = req.Clone();
            temp.ToJSON(ref buf);
            j = Encoding.UTF8.GetString(buf.AsSpan());
        }
        finally
        {
            buf.Dispose();
        }
        rsp.Message = $"success. req={j}";
        return default;
    }
}
