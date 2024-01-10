using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas.Models
{
    public class WorkspaceModel
    {
        public bool? IsRelative { get; set; }
        public string NodeConfPath { get; set; }
        public string WorkDir { get; set; }
    }
}
