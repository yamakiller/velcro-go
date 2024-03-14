using Editor.Framework;
using Editor.Panels.Model;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.ComponentModel;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using System.Windows.Shapes;

namespace Editor.Contrels
{
    /// <summary>
    /// AttributeList.xaml 的交互逻辑
    /// </summary>
    public partial class AttributeList : UserControl, INotifyPropertyChanged
    {
        public AttributeList()
        {
            InitializeComponent();
        }

        public static readonly DependencyProperty AttributeMapProperty =
        DependencyProperty.Register("AttributeMap", typeof(ObservableCollection<Attribute>), typeof(AttributeList));

        public ObservableCollection<Attribute> AttributeMap
        {
            get
            {
                return (ObservableCollection<Attribute>)GetValue(AttributeMapProperty);
            }
            set
            {
                SetValue(AttributeMapProperty, value);
                //OnPropertyChanged("AttributeMap");
            }
        }


        private void ClickDelete(object sender, RoutedEventArgs e)
        {
            var tmp = AttributeMap;
            var model = tmp.FirstOrDefault(t => t.Key == (sender as Button)?.CommandParameter.ToString());
            if (model != null)
            {
                model.PropertyChanged -= AttributePropertyChanged;
                tmp.Remove(model);
                AttributeMap = tmp;
            }
            this.datagrid.ItemsSource = null;
            this.datagrid.ItemsSource = AttributeMap;
        }

        #region INotifyPropertyChanged Members

        public event PropertyChangedEventHandler? PropertyChanged;

        protected void OnPropertyChanged(string name)
        {
            if (PropertyChanged != null)
                PropertyChanged(this, new PropertyChangedEventArgs(name));
        }

        #endregion

        public class Attribute : IEditableObject
        {
            private string _index;
            public string Index
            {
                get
                {
                    return _index;
                }
                set { _index = value; }
            }

            private string _key;
            public string Key
            {
                get
                {
                    return _key;
                }
                set
                {
                    if (_key == value) return;
                    _key = value;
                }
            }
            private string _value;
            public string Value
            {
                get
                {
                    return _value;
                }
                set
                {
                    if (_value == value) return;
                    _value = value;
                }
            }

            #region INotifyPropertyChanged

            public event PropertyChangedEventHandler PropertyChanged;

            private void OnPropertyChanged(string propertyName)
            {
                if (PropertyChanged != null)
                {
                    PropertyChanged(this, new PropertyChangedEventArgs(propertyName));
                }
            }

            #endregion

            public void BeginEdit()
            {

            }
            public void CancelEdit()
            {

            }
            public void EndEdit()
            {
                OnPropertyChanged("Attribute");
            }
        }

        private void Button_Init_Click(object sender, RoutedEventArgs e)
        {
            this.pop.IsOpen = true;
            foreach (Attribute attribute in AttributeMap)
            {
                attribute.PropertyChanged += AttributePropertyChanged;
            }
            this.datagrid.ItemsSource = null;
            this.datagrid.ItemsSource = AttributeMap;
        }

        private void Button_Add_Click(object sender, RoutedEventArgs e)
        {
            var tmp = AttributeMap;
            var model = tmp.FirstOrDefault(t => t.Key == "新增Key");
            if (model == null)
            {
                var att = new Attribute() { Index = Guid.NewGuid().ToString(), Key = "新增Key", Value = "新增Value" };
                att.PropertyChanged += AttributePropertyChanged;
                tmp.Add(att);

                AttributeMap = tmp;
            }

        }

        void AttributePropertyChanged(object sender, PropertyChangedEventArgs e)
        {
            var tmp = AttributeMap;
            var att = sender as Attribute;
            foreach (Attribute attribute in tmp)
            {
                if (attribute.Index == att.Index)
                {
                    attribute.Key = att.Key;
                    attribute.Value = att.Value;
                }
            }
            AttributeMap = tmp;
        }
    }
}
