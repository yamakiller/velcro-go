using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class BehaviorTree
    {
        public required string FileName { get; set; }
        public Files.Behavior3Tree? Tree { get; set; }
    }
}
