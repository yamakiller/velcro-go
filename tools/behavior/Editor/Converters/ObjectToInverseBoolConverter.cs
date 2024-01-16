using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Data;

namespace Editor.Converters
{
    [ValueConversion(typeof(object), typeof(bool))]
    class ObjectToInverseBoolConverter : IValueConverter
    {
        public static readonly ObjectToInverseBoolConverter Instance = new ObjectToInverseBoolConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null)
                return false;

            return true;
        }
        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            return Convert(value, targetType, parameter, culture);
        }

    }
}
