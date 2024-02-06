using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas.Files
{
    public class Behavior3Node
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
        public List<string>? Children { get; set;}
        [JsonProperty(PropertyName = "color")]
        public string? Color {  get; set; }
        [JsonProperty(PropertyName = "properties")]
        public Dictionary<string, object>? Properties { get; set; }
    }


    public class Behavior3Tree
    {
        [JsonProperty(PropertyName = "id")] 
        public required string ID { get; set; }
        [JsonProperty(PropertyName = "title")]
        public required string Title { get; set; }
        [JsonProperty(PropertyName = "description")]
        public required string Description { get; set; }
        [JsonProperty(PropertyName = "properties")]
        public Dictionary<string, object>? Properties { get; set; }
        [JsonProperty(PropertyName = "nodes")]
        public Dictionary<string, Behavior3Node>? Nodes {  get; set; }
    }

}
