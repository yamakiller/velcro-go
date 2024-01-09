using Microsoft.VisualBasic;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class BehaviorNode
    {
        public string id {  get; set; }
        public string name { get; set; }
        public string category {  get; set; } // 类别对应类型
        public string title { get; set; }
        public string description { get; set; }
        public List<BehaviorNode> children {  get; set; }
        public BehaviorNode child { get; set; }
        public Dictionary<string, object>? properties { get; set; }

    }
}
