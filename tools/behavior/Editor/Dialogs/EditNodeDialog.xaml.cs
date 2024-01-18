using Editor.Datas.Models;
using Editor.Framework;
using Editor.Utils;
using Editor.ViewModels;
using Microsoft.Win32;
using System.Collections.ObjectModel;
using System.Windows;
using System.Windows.Controls;


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

        private void CreateArgs_Click(object sender, RoutedEventArgs e)
        {

            if (((EditNodeDialogViewModel)(DataContext)).SelectedNode == null)
            {
                return;
            }

            ArgsTypeEditDialog argsEditDialog = new ArgsTypeEditDialog(newArgName());
            var result = argsEditDialog.ShowDialog();
            if (result != true)
            {
                return;
            }

            ArgsDefType argsType = new ArgsDefType();
            argsType.name = argsEditDialog.Name == null ? "newArgs" : argsEditDialog.Name;
            argsType.type = argsEditDialog.Type == null ? "int" : argsEditDialog.Type;
            argsType.defaultValue = argsEditDialog.DefaultValue;
            argsType.desc = argsEditDialog.Desc;

            if (((EditNodeDialogViewModel)(DataContext)).SelectedNode.args == null)
            {
                ((EditNodeDialogViewModel)(DataContext)).SelectedNode.args = new ObservableCollection<ArgsDefType>();
            }

            ((EditNodeDialogViewModel)DataContext).SelectedNode.args.Add(argsType);
            ((EditNodeDialogViewModel)DataContext).SelectedNodeArgs = ((EditNodeDialogViewModel)DataContext).SelectedNode.args;
        }

        private void DeleteArgs_Click(object sender, RoutedEventArgs e)
        {
            if (((EditNodeDialogViewModel)(DataContext)).SelectedNode == null)
            {
                return;
            }

            if (((EditNodeDialogViewModel)(DataContext)).SelectedNode.args == null)
            {
                return;
            }



            ((EditNodeDialogViewModel)(DataContext)).SelectedNode.args.RemoveAt(ArgsCtrl.SelectedIndex);
            ((EditNodeDialogViewModel)(DataContext)).SelectedNodeArgs = ((EditNodeDialogViewModel)(DataContext)).SelectedNode.args;
            if (((EditNodeDialogViewModel)(DataContext)).SelectedNode.args.Count > 0)
            {
                ArgsCtrl.SelectedIndex = 0;
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

        private string newArgName()
        {
            string newName = "";
            for (int i =1;i < 65535;)
            {
             new_name_label:
                newName = "newArgs_" + i.ToString();
                if (((EditNodeDialogViewModel)(DataContext)).SelectedNode == null ||
                    ((EditNodeDialogViewModel)(DataContext)).SelectedNode.args == null)
                {
                    break;
                }

                foreach(ArgsDefType t in ((EditNodeDialogViewModel)(DataContext)).SelectedNode.args)
                {
                    if (t.name == newName)
                    {
                        i++;
                        goto new_name_label;
                    }
                }
                break;
            }
            return newName;
        }
    }
}
