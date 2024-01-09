using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class Workspace
    {
        // 工作空间名
        public string Name { get; set; }
        // 工作空间描述
        public string Description { get; set; }
        // 行为树表
        public Dictionary<string, BehaviorTree> Trees { get; set; }
        // 工作空间创建时间
        public DateTime CreateTime { get; set; }
        // 工作空间最后修改时间
        public DateTime ModifyTime { get; set; }
    }
}
