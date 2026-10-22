extern alias New;
extern alias Old;

using System.Collections;
using System.Reflection;
using System.Text;
using System.Text.Json.Nodes;
using Google.Protobuf;
using Microsoft.CodeAnalysis;
using Microsoft.CodeAnalysis.CSharp;
using Microsoft.CodeAnalysis.CSharp.Syntax;
using QiWa.Common;

var newer = (New.Generated.Demo.AllTypes)Sample(typeof(New.Generated.Demo.AllTypes), 0)!;
var older = (Old.Generated.Demo.AllTypes)Sample(typeof(Old.Generated.Demo.AllTypes), 0)!;
var newBytes = Encode(newer);
var oldBytes = Encode(older);
if (!newBytes.SequenceEqual(oldBytes)) throw new Exception("old/new protobuf differs");
var newReader = new New.Generated.Demo.ReadonlyAllTypes();
var oldReader = new Old.Generated.Demo.ReadonlyAllTypes();
Check(newReader.FromProtobuf(oldBytes));
Check(oldReader.FromProtobuf(newBytes));
var newClone = new New.Generated.Demo.AllTypes();
var oldClone = new Old.Generated.Demo.AllTypes();
newReader.Clone(ref newClone);
oldReader.Clone(ref oldClone);
EqualJSON(JSON(newer), JSON(newClone));
EqualJSON(JSON(older), JSON(oldClone));
Check(newReader.FromJSON(JSON(older)));
Check(oldReader.FromJSON(JSON(newer)));
newReader.Clone(ref newClone);
oldReader.Clone(ref oldClone);
EqualJSON(JSON(newer), JSON(newClone));
EqualJSON(JSON(older), JSON(oldClone));
var oracle = Generated.Demo.GrpcProto.AllTypes.Parser.ParseFrom(newBytes);
Check(newReader.FromProtobuf(oracle.ToByteArray()));
newReader.Clone(ref newClone);
EqualJSON(JSON(newer), JSON(newClone));
var beforeReset = JSON(newClone);
newReader.Reset();
Array.Fill(newBytes, (byte)0xFF);
EqualJSON(beforeReset, JSON(newClone));
Console.WriteLine("AllTypes: old/new JSON+Protobuf, Google.Protobuf oracle, Clone/Reset passed");

if (args.Length == 2 && args[0] == "--vectors")
{
    var root = args[1];
    var expected = File.ReadAllBytes(Path.Combine(root, "go.json"));
    var binary = File.ReadAllBytes(Path.Combine(root, "go.pb"));
    newReader.Reset();
    Check(newReader.FromProtobuf(binary));
    newReader.Clone(ref newClone);
    newReader.Reset();
    Array.Fill(binary, (byte)0xFF);
    EqualCrossLanguageJSON(expected, JSON(newClone));
    File.WriteAllBytes(Path.Combine(root, "csharp.pb"), Encode(newClone));
    File.WriteAllBytes(Path.Combine(root, "csharp.json"), JSON(newClone));
    var official = Generated.Demo.GrpcProto.AllTypes.Parser.ParseFrom(Encode(newClone));
    File.WriteAllBytes(Path.Combine(root, "oracle.pb"), official.ToByteArray());
    Check(newReader.FromJSON(expected));
    newReader.Clone(ref newClone);
    EqualCrossLanguageJSON(expected, JSON(newClone));
    Console.WriteLine("Go AllTypes protobuf/JSON vectors decoded; C#/official vectors written");
    var boundary = File.ReadAllBytes(Path.Combine(root, "go-unsigned-map.pb"));
    newReader.Reset();
    oldReader.Reset();
    Check(newReader.FromProtobuf(boundary));
    Check(oldReader.FromProtobuf(boundary));
    newReader.Clone(ref newClone);
    oldReader.Clone(ref oldClone);
    if (newClone.M4[ulong.MaxValue] != 4 || oldClone.M4[ulong.MaxValue] != 4 || newClone.Fixed64Value != ulong.MaxValue || oldClone.Fixed64Value != ulong.MaxValue)
        throw new Exception("uint64 map key protobuf decoding changed");
    EqualJSON(JSON(oldClone), JSON(newClone));
    var boundaryJSON = JsonNode.Parse(JSON(newClone))!;
    if (boundaryJSON["m4"]?["-1"]?.GetValue<int>() != 4 || boundaryJSON["Fixed64Value"]?.GetValue<string>() != "-1")
        throw new Exception("legacy uint64 map JSON behavior changed; review boundary expectation");
    Console.WriteLine("KNOWN BASELINE LIMITATION: C# JSON casts uint64 map keys/fixed64 above Int64.MaxValue to signed; old/new identical, protobuf correct");
    newReader.Reset();
    oldReader.Reset();
    var boundaryInput = File.ReadAllBytes(Path.Combine(root, "go-unsigned.json"));
    if (!newReader.FromJSON(boundaryInput).Err() || !oldReader.FromJSON(boundaryInput).Err())
        throw new Exception("legacy unsigned JSON decoding behavior changed; review boundary expectation");
    Console.WriteLine("KNOWN BASELINE LIMITATION: C# JSON decoding rejects upper unsigned range; old/new both reject");
}
else if (args.Length == 2)
{
    var baseline = Declarations(args[0]);
    var migrated = Declarations(args[1]);
    if (!baseline.Keys.Order().SequenceEqual(migrated.Keys.Order())) throw new Exception("C# declaration inventory differs");
    foreach (var entry in baseline)
        if (entry.Value != migrated[entry.Key]) throw new Exception("C# declaration changed: " + entry.Key);
    Console.WriteLine($"C# syntax tokens: {baseline.Count} declarations match");
}

static void Check(Error error)
{
    if (error.Err()) throw new Exception(error.ToString());
}

static byte[] Encode<Type>(Type value) where Type : struct, IEncoder
{
    var buffer = new RentedBuffer(4096);
    try { Check(value.ToProtobuf(ref buffer)); return buffer.AsSpan().ToArray(); }
    finally { buffer.Dispose(); }
}

static byte[] JSON<Type>(Type value) where Type : struct, IEncoder
{
    var buffer = new RentedBuffer(4096);
    try { value.ToJSON(ref buffer); return buffer.AsSpan().ToArray(); }
    finally { buffer.Dispose(); }
}

static void EqualJSON(byte[] expected, byte[] actual)
{
    if (!JsonNode.DeepEquals(JsonNode.Parse(expected), JsonNode.Parse(actual)))
        throw new Exception("JSON values differ: " + Difference(JsonNode.Parse(expected), JsonNode.Parse(actual), "$"));
}

static string Difference(JsonNode? expected, JsonNode? actual, string path)
{
    if (expected is JsonObject left && actual is JsonObject right)
    {
        foreach (var key in left.Select(entry => entry.Key).Union(right.Select(entry => entry.Key)))
            if (!JsonNode.DeepEquals(left[key], right[key])) return Difference(left[key], right[key], path + "." + key);
    }
    return $"{path}: expected={expected}, actual={actual}";
}

static void EqualCrossLanguageJSON(byte[] expected, byte[] actual)
{
    var left = JsonNode.Parse(expected);
    var right = JsonNode.Parse(actual);
    NormalizeIntegers(left);
    NormalizeIntegers(right);
    if (!JsonNode.DeepEquals(left, right))
        throw new Exception("cross-language values differ: " + Difference(left, right, "$"));
}

static void NormalizeIntegers(JsonNode? node)
{
    if (node is JsonObject fields)
    {
        foreach (var key in fields.Select(field => field.Key).ToArray())
        {
            if (key is "Int64Value" or "Uint64Value" or "Sint64Value" or "Fixed64Value" or "Sfixed64Value")
                fields[key] = JsonValue.Create(fields[key]!.ToString());
            else NormalizeIntegers(fields[key]);
        }
    }
    else if (node is JsonArray elements)
        foreach (var element in elements) NormalizeIntegers(element);
}

static object? Sample(Type type, int depth)
{
    if (type == typeof(string)) return @"hello ""世界"" backslash \\" + Environment.NewLine;
    if (type == typeof(byte[])) return new byte[] { 0, 1, 127, 255 };
    if (type == typeof(bool)) return true;
    if (type.IsEnum) return Enum.ToObject(type, 1);
    if (type == typeof(decimal)) return 1.12345m;
    if (type == typeof(float)) return 1.5f;
    if (type == typeof(double)) return 1.5;
    if (type.IsPrimitive) return Convert.ChangeType(42, type);
    if (depth >= 5 && type.IsClass && type.Name.EndsWith("Wrapper")) return null;
    if (depth > 6) return type.IsValueType ? Activator.CreateInstance(type) : null;
    var result = Activator.CreateInstance(type)!;
    if (result is IDictionary dictionary)
    {
        var arguments = type.GetGenericArguments();
        var value = depth < 5 ? Sample(arguments[1], depth + 1) : null;
        if (value != null) dictionary.Add(Sample(arguments[0], depth + 1)!, value);
    }
    else if (result is IList list)
        list.Add(Sample(type.GetGenericArguments()[0], depth + 1));
    else
        foreach (var field in type.GetFields(BindingFlags.Instance | BindingFlags.Public))
            field.SetValue(result, Sample(field.FieldType, depth + 1));
    return result;
}

static SortedDictionary<string, string> Declarations(string root)
{
    var result = new SortedDictionary<string, string>();
    foreach (var path in Directory.GetFiles(root, "*.cs"))
    {
        var tree = CSharpSyntaxTree.ParseText(File.ReadAllText(path));
        foreach (var type in tree.GetRoot().DescendantNodes().OfType<BaseTypeDeclarationSyntax>())
        {
            if (type is EnumDeclarationSyntax enumeration)
            {
                result.Add(type.Identifier.Text, Tokens(enumeration));
                continue;
            }
            if (type is not TypeDeclarationSyntax declaration) continue;
            if (declaration.BaseList != null) result.Add(type.Identifier.Text + ":bases", Tokens(declaration.BaseList));
            foreach (var member in declaration.Members)
            {
                string name = member switch
                {
                    MethodDeclarationSyntax method => method.Identifier.Text + Tokens(method.ParameterList),
                    FieldDeclarationSyntax field => string.Join(",", field.Declaration.Variables.Select(variable => variable.Identifier.Text)),
                    PropertyDeclarationSyntax property => property.Identifier.Text,
                    _ => throw new Exception("unhandled declaration " + member.Kind())
                };
                result.Add(type.Identifier.Text + "." + name, Tokens(member));
            }
        }
    }
    return result;
}

static string Tokens(SyntaxNode node) => string.Join(" ", node.DescendantTokens().Select(token => token.Text));
