using Editor.Framework;
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
using System.Windows.Shapes;

namespace Editor.Dialogs
{
    /// <summary>
    /// WhatDialog.xaml 的交互逻辑
    /// </summary>
    public partial class WhatDialog : Window
    {
        public static bool ShowWhatMessage(string caption , string message )
        {
            WhatDialog wdlg = new WhatDialog();
            wdlg.Caption = caption;
            wdlg.Message = message;

            var result = wdlg.ShowDialog();

            return result == true ? true : false;
        }
        public WhatDialog()
        {
            InitializeComponent();
            DataContext = new WhatDialogViewModel();
        }

        public string Caption
        {
            set
            {
                ((WhatDialogViewModel)DataContext).Caption = value;
            }
        }

        public string Message
        {
            set
            {
                ((WhatDialogViewModel)DataContext).Message = value;
            }
        }

        private void Accept_Click(object sender, RoutedEventArgs e)
        {
            this.DialogResult = true;
            this.Close();
        }

        private void Cancel_Click(object sender, RoutedEventArgs e)
        {
            this.DialogResult = false;
            this.Close();
        }
    }
}
