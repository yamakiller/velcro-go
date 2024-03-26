using Confluent.Kafka;
using Editor.Datas;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.ComponentModel;
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
    public class Tree
    {
        public string ID { get; set; }
        public string Name { get; set; }
    }
    /// <summary>
    /// SelectedRootTreeDialog.xaml 的交互逻辑
    /// </summary>
    public partial class SelectedRootTreeDialog : Window
    {
        public SelectedRootTreeDialog()
        {
            WindowStartupLocation = WindowStartupLocation.CenterScreen;
            InitializeComponent();
            DataContext = this;
        }
        public List<Tree> _SelectedTrees = new List<Tree>();
        public List<Tree> SelectedTrees { get {
                return _SelectedTrees;
            } private set {
                _SelectedTrees = value;
            } }
        public Tree? SelectedRoot = null;

        public bool? ShowTrees(ObservableCollection<BehaviorTree> trees)
        {
            SelectedTrees.Clear();
            List<Tree> list = new List<Tree>();
            foreach (var item in trees)
            {
                list.Add(new Tree() { ID = item.ID,Name = item.Title});
            }
            SelectedTrees = list;
            SelectedTreeDataGrid.ItemsSource = null;
            SelectedTreeDataGrid.ItemsSource = SelectedTrees;
            return ShowDialog();
        }

        public void onSelectRootButton(object sender, RoutedEventArgs e)
        {
            var root = SelectedTreeDataGrid.SelectedValue as Tree;
            if(root == null)
            {
                MessageBox.Show("未选择默认行为树");
                return;
            }

            SelectedRoot = root;

            this.DialogResult = true;
            this.Close();
        }
    }
}
