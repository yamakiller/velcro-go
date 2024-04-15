using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Data;

namespace Editor.Converters
{

    [ValueConversion(typeof(bool), typeof(bool))]
    class DebugToEnabledConverter : IValueConverter
    {
        public static readonly DebugToEnabledConverter Instance = new DebugToEnabledConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null)
                return false;

            return (bool)value == false;
        }
        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            return Convert(value, targetType, parameter, culture);
        }
    }
}
