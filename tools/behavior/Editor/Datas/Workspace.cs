using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class Workspace
    {
        public required string Name { get; set; }
        public required string Dir { get; set; }
        public required ObservableCollection<BehaviorTree> Trees { get; set; }
    }
}
