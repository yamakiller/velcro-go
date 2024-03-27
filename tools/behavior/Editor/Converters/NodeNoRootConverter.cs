using Behavior.Diagrams.Controls;
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
    class NodeNoRootConverter : IValueConverter
    {
        public static readonly NodeNoRootConverter Instance = new NodeNoRootConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null || value == "")
                return true;
            return NodeKindConvert.ToKind(value as string) != NodeKinds.Root;
        }
        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            return Convert(value, targetType, parameter, culture);
        }
    }
}
