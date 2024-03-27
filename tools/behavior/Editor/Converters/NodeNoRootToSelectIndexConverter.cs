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

    [ValueConversion(typeof(string), typeof(int))]
    class NodeNoRootToSelectIndexConverter : IValueConverter
    {
        public static readonly NodeNoRootToSelectIndexConverter Instance = new NodeNoRootToSelectIndexConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null || value == "")
                return -1;

            return (int)NodeKindConvert.ToKind(value as string)-1;
        }
        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null)
                return -1;
            return NodeKindConvert.ToCategory((NodeKinds)value + 1);
        }
    }
}
