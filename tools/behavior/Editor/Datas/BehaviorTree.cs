using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class BehaviorTree
    {
        public string FilePath {  get; set; }
        public string Sha1 { get; set; }
        public Models.BehaviorTreeModel TreeModel { get; set; }
    }
}
