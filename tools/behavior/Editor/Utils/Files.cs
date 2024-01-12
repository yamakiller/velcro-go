using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace Editor.Utils
{
    public class Files
    {
        public static string WriteBehaviorTree(Datas.BehaviorTree behaviorTree)
        {
            var options = new JsonSerializerOptions { WriteIndented = true };
            var jsonContent = JsonSerializer.Serialize<Datas.Models.BehaviorTreeModel>(behaviorTree.TreeModel, options);
            if (!File.Exists(behaviorTree.FilePath))
            {
                File.WriteAllText(behaviorTree.FilePath, jsonContent);
            }
            else
            {
                FileStream fs = File.OpenWrite(behaviorTree.FilePath);
                fs.Seek(0, SeekOrigin.Begin);
                fs.SetLength(0);
                fs.Write(Encoding.UTF8.GetBytes(jsonContent));
                fs.Close();
            }

            return jsonContent;
        }
        public static string? AutoFileNameNumber(string path, string prefix, string suffix)
        {
            for (int i=1;i < 65535;i++)
            {
                string filename = prefix + "_" + i.ToString() + suffix;
                string filepath = Path.Combine(path, filename);
                if (!File.Exists(filepath))
                {
                    return filename;
                }
            }

            return null;
        }
        public static void ScanFiles(string directory, Action<string> cb)
        {
            DirectoryInfo di = new DirectoryInfo(directory);
            findFile(di, cb);
        }
        static void findFile(DirectoryInfo di, Action<string> cb)
        {
            if (di.Exists)
            {

            }

            FileInfo[] files = di.GetFiles();
            for (int i = 0; i < files.Length; i++)
            {
                cb.Invoke(files[i].FullName);
            }

            DirectoryInfo[] dirs = di.GetDirectories();
            for (int i = 0; i < dirs.Length; i++)
            {
                cb.Invoke(dirs[i].FullName);
                findFile(dirs[i], cb);
            }
   
        }
    }
}
