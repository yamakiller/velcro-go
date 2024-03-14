using Editor.Panels.Model;
using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Data;

namespace Editor.Converters
{
    [ValueConversion(typeof(string), typeof(bool))]
    class NodeNoRootToVisibilityConverter : IValueConverter
    {
        public static readonly NodeNoRootToVisibilityConverter Instance = new NodeNoRootToVisibilityConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null || value == "")
                return "Collapsed";

            if (NodeKindConvert.ToKind(value as string) != NodeKinds.Root)
            {
                return "Visible";
            }
            return "Collapsed";
        }
        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            return Convert(value, targetType, parameter, culture);
        }
    }
}
