using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Utils
{
    public class ShortGuid
    {
        private string m_shortid;
        public ShortGuid(Guid newGuid) 
        {
            m_shortid = toShortGuid(newGuid);
        }

        public ShortGuid() 
        {
            m_shortid = toShortGuid(Guid.NewGuid());
        }

        public string ToString() { return m_shortid; }
        private string toShortGuid(this Guid newGuid)
        {
            string modifiedBase64 = Convert.ToBase64String(newGuid.ToByteArray())
                .Replace('+', '-').Replace('/', '_')
                .Substring(0, 22);
            return modifiedBase64;
        }
    }
   
}
