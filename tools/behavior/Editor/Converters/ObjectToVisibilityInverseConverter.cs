using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Data;

namespace Editor.Converters
{
    [ValueConversion(typeof(object), typeof(string))]
    class ObjectToVisibilityInverseConverter : IValueConverter
    {
        public static readonly ObjectToVisibilityInverseConverter Instance = new ObjectToVisibilityInverseConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null)
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
