using Microsoft.Win32;
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

namespace Editor.Dialogs
{
    /// <summary>
    /// CreateWorkspaceDialog.xaml 的交互逻辑
    /// </summary>
    public partial class CreateWorkspaceDialog : Window
    {
        public CreateWorkspaceDialog()
        {
            InitializeComponent();
            WorkspaceNameInput.Text = "NewWorkspace";
        }

        public string WorkspaceName
        {
            get
            {
                return WorkspaceNameInput.Text;
            }
        }

        public string WorkspaceFolder
        {
            get
            {
                return WorkspaceFolderInput.Text;
            }
        }

        private void Cancel_Click(object sender, RoutedEventArgs e)
        {
            this.DialogResult = false;
            this.Close();
        }

        private void Applay_Click(object sender, RoutedEventArgs e)
        {
            if (string.IsNullOrEmpty(WorkspaceNameInput.Text))
            {
                MessageBox.Show("请输入空间名称");
                return;
            }

            if (string.IsNullOrEmpty(WorkspaceFolderInput.Text))
            {
                MessageBox.Show("请输入工作目录");
                return;
            }


            this.DialogResult = true;
            this.Close();
        }

        private void Folder_Click(object sender, RoutedEventArgs e)
        {
            var folderDialog = new OpenFolderDialog()
            {
                Title = "Workspace",
                InitialDirectory = Environment.GetFolderPath(Environment.SpecialFolder.Personal),
                Multiselect = false,
            };
            var result = folderDialog.ShowDialog();
            if (result != true)
            {
                return;
            }

            WorkspaceFolderInput.Text = folderDialog.FolderName;
        }
    }
}
