using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas.Files
{
    public class B3Node
    {
        [JsonProperty(PropertyName = "id")]
        public required string ID { get; set; }
        [JsonProperty(PropertyName = "name")]
        public required string Name { get; set; }
        [JsonProperty(PropertyName = "category")]
        public required string Category { get; set; }
        [JsonProperty(PropertyName = "title")]
        public required string Title { get; set; }
        [JsonProperty(PropertyName = "description")]
        public required string Description { get; set; }
        [JsonProperty(PropertyName = "children")]
        public List<string>? Children { get; set; }
        [JsonProperty(PropertyName = "child")]
        public string Child { get; set; }
        [JsonProperty(PropertyName = "properties")]
        public Dictionary<string, object>? Properties { get; set; }
    }
    public class B3Tree
    {
        [JsonProperty(PropertyName = "id")]
        public required string ID { get; set; }
        [JsonProperty(PropertyName = "title")]
        public required string Title { get; set; }
        [JsonProperty(PropertyName = "description")]
        public required string Description { get; set; }
        [JsonProperty(PropertyName = "root")]
        public  string? Root { get; set; }
        [JsonProperty(PropertyName = "properties")]
        public  Dictionary<string, object>? Properties { get; set; }
        [JsonProperty(PropertyName = "nodes")]
        public  Dictionary<string, B3Node>? Nodes { get; set; }
    }
    public class B3Project
    {
        [JsonProperty(PropertyName = "selectedTree")]
        public required string SelectedTree { get; set; }
        [JsonProperty(PropertyName = "scope")]
        public required string Scope { get; set; }
        [JsonProperty(PropertyName = "trees")]
        public required B3Tree[]? Trees { get; set; }
    }
    public class B3File
    {
        [JsonProperty(PropertyName = "name")]
        public required string Name { get; set; }
        [JsonProperty(PropertyName = "data")]
        public required B3Project Data { get; set; }
    }
}
