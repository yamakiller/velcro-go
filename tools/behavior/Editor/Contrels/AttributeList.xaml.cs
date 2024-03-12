using System;
using System.Collections.Generic;
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
    public partial class AttributeList : UserControl
    {
        public AttributeList()
        {
            InitializeComponent();
        }
        public static readonly DependencyProperty AttributeMapProperty =
        DependencyProperty.Register("AttributeMap", typeof(Dictionary<string, object>), typeof(AttributeList));
        public Dictionary<string, object> AttributeMap
        {

            get => (Dictionary<string, object>)GetValue(AttributeMapProperty);
            set {
                SetValue(AttributeMapProperty, value);
            }
        }

        private void btn_Click(object sender, RoutedEventArgs e)
        {
            pop.IsOpen = true;
            AttributeMap.Clear();
            AttributeMap.Add("ceshi1","测试1");
            AttributeMap.Add("ceshi2", "测试2");
            AttributeMap.Add("ceshi3", "测试3");
        }
    }
}
