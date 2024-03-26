using System;
using System.Collections;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Globalization;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Data;
using System.Windows.Media;

namespace Editor.Converters
{
    [ValueConversion(typeof(ObservableCollection<Contrels.AttributeList.Attribute>), typeof(Dictionary<string, KeyValuePair<string, object>>))]
    class MapToObservableConverter : IValueConverter
    {
        public static readonly MapToObservableConverter Instance = new MapToObservableConverter();

        public object Convert(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null || value.ToString() == "")
            {
                return new ObservableCollection<Contrels.AttributeList.Attribute>();
            }
            ObservableCollection<Contrels.AttributeList.Attribute> newValue = new ObservableCollection<Contrels.AttributeList.Attribute>();
            foreach (var attribute in value as Dictionary<string, KeyValuePair<string,object>> )
            {
                newValue.Add(new Contrels.AttributeList.Attribute()
                {
                    Index = attribute.Key,
                    Key = attribute.Value.Key,
                    Value = attribute.Value.Value,
                });
            }
            return newValue;
        }

        public object ConvertBack(object value, Type targetType, object parameter, CultureInfo culture)
        {
            if (value == null || value.ToString() == "")
            {
                return new Dictionary<string, KeyValuePair<string, object>>();
            }

            Dictionary<string, KeyValuePair<string, object>> newValue = new Dictionary<string, KeyValuePair<string, object>>();
            foreach (var attribute in value as ObservableCollection<Contrels.AttributeList.Attribute>)
            {
                newValue.Add(attribute.Index,new KeyValuePair<string,object>(attribute.Key,attribute.Value));
            }

            return newValue;
        }
    }
}
