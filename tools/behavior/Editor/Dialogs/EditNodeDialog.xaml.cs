using Editor.Framework;
using Editor.Utils;
using Editor.ViewModels;
using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Threading.Tasks.Sources;
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

        public Datas.Models.BehaviorNodeTypeModel[]? GetTypes()
        {
            return m_types.ToArray();
        }

        /// <summary>
        /// 选择文件夹
        /// </summary>
        /// <param name="sender"></param>
        /// <param name="e"></param>
        private void SourceFileSelected_Click(object sender, RoutedEventArgs e)
        {
            OpenFileDialog openFileDlg = new OpenFileDialog();
            openFileDlg.Multiselect = false;
            openFileDlg.Filter = "Behavior tree node source code file (.go)| *.go";
            if (openFileDlg.ShowDialog() != true)
            {
                return;
            }

        
            ((EditNodeDialogViewModel)(DataContext)).SelectedNodeFile = openFileDlg.FileName;
        }

        private void CreateNode_Click(object sender, RoutedEventArgs e)
        {
            var nodeNew = Configs.Config.UnknowNodeType();
            m_types.Add(nodeNew);
            if (((EditNodeDialogViewModel)(DataContext)).SelectedNode == null)
            {
                ((EditNodeDialogViewModel)(DataContext)).SelectedNode = m_types[0];
            }
        }

        private void DeleteNode_Click(object sender, RoutedEventArgs e)
        {
            if (((EditNodeDialogViewModel)(DataContext)).SelectedNode == null)
            {
                return;
            }

            for (int i=0;i< m_types.Count;i++)
            {
                if (m_types[i] == ((EditNodeDialogViewModel)(DataContext)).SelectedNode)
                {
                    m_types.RemoveAt(i);
                    if (m_types.Count == 0)
                    {
                        ((EditNodeDialogViewModel)(DataContext)).SelectedNode = null;
                    }
                    else if (i >= m_types.Count)
                    {
                        ((EditNodeDialogViewModel)(DataContext)).SelectedNode = m_types[i - 1];
                    }
                    else
                    {
                        ((EditNodeDialogViewModel)(DataContext)).SelectedNode = m_types[i];
                    }
                    break;
                }
            }
        }

        private void Applay_Click(object sender, RoutedEventArgs e)
        {
            this.DialogResult = true;
            this.Close();
        }

        private void Cancel_Click(object sender, RoutedEventArgs e)
        {
            this.DialogResult = false;
            this.Close();
        }

        private void TreeView_SelectedItemChanged(object sender, RoutedPropertyChangedEventArgs<object> e)
        {
            ((EditNodeDialogViewModel)(DataContext)).SelectedNode = (Datas.Models.BehaviorNodeTypeModel)e.NewValue;
            ((EditNodeDialogViewModel)(DataContext)).SelectedNodeFile = ((EditNodeDialogViewModel)(DataContext)).SelectedNode.src;
        }

        private void Window_Loaded(object sender, RoutedEventArgs e)
        {
            if (m_types.Count > 0)
            {
                ((EditNodeDialogViewModel)(DataContext)).SelectedNode = m_types[0];
                if (NodeTreeView.Items.Count > 0)
                {
                    TreeViewItem childContainer = NodeTreeView.ItemContainerGenerator.ContainerFromItem(NodeTreeView.Items[0]) as TreeViewItem;
                    childContainer.IsSelected = true;
                }
            }
        }
    }
}
