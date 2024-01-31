﻿
using System.Windows.Data;

namespace Bga.Diagrams.Utils
{
    [System.Diagnostics.CodeAnalysis.SuppressMessage("Microsoft.Performance", "CA1812:AvoidUninstantiatedInternalClasses")]
    internal class DebuggingConverter : IValueConverter
    {
        public static DebuggingConverter Instance
        {
            get { return new DebuggingConverter(); }
        }

        public object Convert(object value, Type targetType, object parameter, System.Globalization.CultureInfo culture)
        {
            return value;
        }

        public object ConvertBack(object value, Type targetType, object parameter, System.Globalization.CultureInfo culture)
        {
            return value;
        }
    }
}