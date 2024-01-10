using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Utils
{
    public static  class ShortGuid
    {
        public static string Next()
        {
            return toShortGuid(Guid.NewGuid());
        }
        private static string toShortGuid(this Guid newGuid)
        {
            string modifiedBase64 = Convert.ToBase64String(newGuid.ToByteArray())
                .Replace('+', '-').Replace('/', '_')
                .Substring(0, 22);
            return modifiedBase64;
        }
    }
   
}
