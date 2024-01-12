using Editor.Framework;
using Editor.ViewModels;
using Microsoft.Win32;
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
using System.Windows.Shapes;

namespace Editor.Dialogs
{
    /// <summary>
    /// EditNodeDialog.xaml 的交互逻辑
    /// </summary>
    public partial class EditNodeDialog : Window
    {
        private ObservableCollection<Datas.Models.BehaviorNodeTypeModel> m_types = new ObservableCollection<Datas.Models.BehaviorNodeTypeModel>();
        public EditNodeDialog(Datas.Models.BehaviorNodeTypeModel[] types)
        {
            InitializeComponent();
            this.DataContext = new EditNodeDialogViewModel(m_types);
            foreach (var t in types)
            {
                m_types.Add(t);
            }
        }

        /// <summary>
        /// 选择文件夹
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void FolderSelected_Click(object sender, RoutedEventArgs e)
        {
            var folderDialog = new OpenFolderDialog()
            {
                Title = "选择代码生成位置",
                InitialDirectory = Environment.GetFolderPath(Environment.SpecialFolder.Personal),
                Multiselect = false,
            };
            var result = folderDialog.ShowDialog();
            if (result != true)
            {
                return;
            }
            ((EditNodeDialogViewModel)(DataContext)).CodeFolder = folderDialog.FolderName;
        }
    }
}
