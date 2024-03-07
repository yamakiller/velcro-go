using System;
using System.Collections.Generic;
using System.Globalization;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Data;
using System.Windows;
using System.Windows.Media;

namespace Editor.Converters
{

    [ValueConversion(typeof(string), typeof(SolidColorBrush))]
    class StringToSolidColorBrushConverter : IValueConverter
    {
        public static readonly StringToSolidColorBrushConverter Instance = new StringToSolidColorBrushConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null || value.ToString() == "")
            {
                return new SolidColorBrush(Colors.Black);
            }
            
            return new SolidColorBrush((Color)ColorConverter.ConvertFromString(value.ToString()));
        }

        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            return Convert(value, targetType, parameter, culture);
        }
    }

}
