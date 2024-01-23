
namespace Editor.Datas
{
    internal class JsonPropertyAttribute : Attribute
    {
        public required string PropertyName { get; set; }
    }
}