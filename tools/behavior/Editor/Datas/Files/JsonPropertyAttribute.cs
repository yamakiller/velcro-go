
namespace Editor.Datas.Files
{
    internal class JsonPropertyAttribute : Attribute
    {
        public required string PropertyName { get; set; }
    }
}