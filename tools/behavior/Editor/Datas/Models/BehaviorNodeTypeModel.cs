using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas.Models
{
    public class ArgsOptional
    {
        public string name { get; set; }
        public object value { get; set; }
    }

    public class ArgsDefType
    {
        public string name { get; set; }
        public string type { get; set; }
        public string desc { get; set; }
        public string defa { get; set; }
        public ArgsOptional[] options { get; set; }
    }

    public class BehaviorNodeTypeModel
    {
        public string name { get; set; }
        public string? type { get; set; }
        public string? desc { get; set; }
        public ArgsDefType[]? args {  get; set; } 
        public string? input {  get; set; }
        public string? output { get; set; }
        public string? src { get; set; }
    }

    public class BehaviorNodeModel
    { 
        public string id {  get; set; }
        public string name { get; set; }
        public string? desc { get; set; }

        public Dictionary<string, object>? args { get; set; }

        public string[]? input { get; set; }
        public string[]? output { get; set; }
        public BehaviorNodeModel[]? children { get; set; }
        public bool? debug { get; set; }
    }

    public class BehaviorTreeModel
    {
        public string name { get; set; }
        public string? desc { get; set; }
        public BehaviorNodeModel root { get; set; }
    }
}
