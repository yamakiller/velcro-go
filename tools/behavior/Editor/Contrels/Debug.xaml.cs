using MaterialDesignThemes.Wpf;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
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
    /// Debug.xaml 的交互逻辑
    /// </summary>
    public partial class Debug : UserControl
    {
        public Debug()
        {
            InitializeComponent();
        }


        public static readonly DependencyProperty KakfaAddressProperty =
        DependencyProperty.Register("KakfaAddress", typeof(string), typeof(Debug));

        public string KakfaAddress
        {
            get { return (string)GetValue(KakfaAddressProperty); }
            set { SetValue(KakfaAddressProperty, value); }
        }

        public void Button_Init_Click(object sender, RoutedEventArgs e)
        {
            this.pop.IsOpen = true;
            PackIconKind iconKind = PackIconKind.None;
        }
    }
}
