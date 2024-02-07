using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Utils
{
    public static class RandFileName
    {
        public static string GetRandName(string dir, List<string>? files, string prefix, string suffix="") 
        {
            for (int i = 1; i < UInt16.MaxValue; i++) 
            { 
                string filePath = Path.Combine(dir, prefix + i + suffix);
                if (File.Exists(filePath))
                {
                    continue;
                } 
                else if (files != null && files.Count > 0)
                {
                    if (files.Find(x => x.Equals(prefix + i + suffix)) != null)
                    {
                        continue;
                    }
                }
                return prefix + i + suffix;
            }

            return string.Empty;
        }
    }
}
