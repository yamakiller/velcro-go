using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using Editor.ViewModels;
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
    /// ArgsTypeEditDialog.xaml 的交互逻辑
    /// </summary>
    public partial class ArgsTypeEditDialog : Window
    {
        public string? Name
        { 
            get { return ((ArgsTypeEditDialogViewModel)(DataContext)).Name; }
        }

        public string? Type
        {
            get { return ((ArgsTypeEditDialogViewModel)(DataContext)).Type; }
        }

        public object? DefaultValue
        {
            get { return ((ArgsTypeEditDialogViewModel)(DataContext)).DefaultValue; }
        }

        public string? Desc
        {
            get { return ((ArgsTypeEditDialogViewModel)(DataContext)).Desc; }
        }

        public ArgsTypeEditDialog(string name)
        {
            InitializeComponent();
            this.DataContext = new ArgsTypeEditDialogViewModel();
            ((ArgsTypeEditDialogViewModel)(DataContext)).Name = name;
            ((ArgsTypeEditDialogViewModel)(DataContext)).Type = "int";
            ((ArgsTypeEditDialogViewModel)(DataContext)).DefaultValue = 0;
        }



        private void Applay_Click(object sender, RoutedEventArgs e)
        {
            if (string.IsNullOrEmpty(((ArgsTypeEditDialogViewModel)(DataContext)).Name))
            {
                ((ArgsTypeEditDialogViewModel)(DataContext)).NameError = "请输入参数名";
                return;
            }

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
