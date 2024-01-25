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
        public static string GetRandName(string dir, string prefix, string suffix="") 
        {
            for (int i = 1; i < UInt16.MaxValue; i++) 
            { 
                string filePath = Path.Combine(dir, prefix + i + suffix);
                if (File.Exists(filePath))
                {
                    continue;
                }
                return prefix + i + suffix;
            }

            return string.Empty;
        }
    }
}
