using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class WorkspaceData
    {
        [JsonProperty(PropertyName = "dir")]
        // 工作目录
        public required string Dir {  get; set; }
        // 包含的树文件
        [JsonProperty(PropertyName = "files")]
        public string[]? Files { get; set; }
    

    }
}
