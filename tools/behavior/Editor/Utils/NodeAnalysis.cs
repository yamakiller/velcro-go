using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Utils
{
    public class NodeAnalysis
    {
       
        // 节点名
        private string name;
        public string Name { get { return name; } }

        // 包/命名空间名
        private string package;
        public string Package {  get { return package; } }
        // 文件编码内容
        private string content;
        public string Content { get { return content; } }
        // 源码文件及路径
        private string filepath;
        public string FilePath { get { return  filepath; } }

        public NodeAnalysis(string filepath) 
        {
            this.filepath = filepath;
        }

        public bool Do()
        {
            name = Path.GetFileNameWithoutExtension(this.filepath);
            var fs = File.OpenText(this.filepath);

            while(true)
            {
                string? sline = fs.ReadLine();
                if (sline == null)
                {
                    break;
                }

                if (sline.IndexOf("package") >= 0)
                {
                    string[] packs = sline.Split(" ");
                    if (packs.Count() >= 2)
                    {
                        package = packs[1];
                    }
                }

                content += sline + "\n";
            }
            fs.Close();
            return true;
        }
    }
}
