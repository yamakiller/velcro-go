using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Utils
{
    public class Files
    {
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
