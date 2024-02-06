using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Media;

namespace Behavior.Diagrams.Utils
{
    public static class ColorHelper
    {
        public static Color ToColor(string colorName)
        {
            if (colorName.StartsWith("#"))
                colorName = colorName.Replace("#", string.Empty);
            int v = int.Parse(colorName, System.Globalization.NumberStyles.HexNumber);
            return new Color()
            {
                A = Convert.ToByte((v >> 24) & 255),
                R = Convert.ToByte((v >> 16) & 255),
                G = Convert.ToByte((v >> 8) & 255),
                B = Convert.ToByte((v >> 0) & 255),
            };
        }
    }
}
