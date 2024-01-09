using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class BehaviorTree
    {
        public string id { get; set; }
        public string title {  get; set; }
        public string description { get; set; }
        // 树的属性
        public Dictionary<string, object>? properties { get; set; }
        // 树所有的节点
        public Dictionary<string, BehaviorNode> nodes {  get; set; }
    }
}
