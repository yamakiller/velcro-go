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
    /// EditorWorkspace.xaml 的交互逻辑
    /// </summary>
    public partial class EditorWorkspace : UserControl
    {
        private TextBlock? tempTextBlock = null;
        public EditorWorkspace()
        {
            InitializeComponent();
        }

        private void TreeView_PreviewMouseRightButtonDown(object sender, MouseButtonEventArgs e)
        {
            DependencyObject source = e.OriginalSource as DependencyObject;
            while(source != null && source.GetType() != typeof(TreeViewItem))
                source = System.Windows.Media.VisualTreeHelper.GetParent(source);
            if (source == null)
            {
                tView.ContextMenu = null;
                return;
            }

            TreeViewItem item = source as TreeViewItem;
            if (!(item.DataContext is Datas.BehaviorTree))
            {
                tView.ContextMenu = null;
                return;
            }

            CreateContextMenu(item);
        }

        private void CreateContextMenu(object positionItem)
        {
          
            ContextMenu contextMenu = new ContextMenu();
            MenuItem menuItem = new MenuItem();
            menuItem.Header = "打开视图";
            // TODO: 加入命令
            contextMenu.Items.Add(menuItem);
            tView.ContextMenu = contextMenu;
        }

        private void TreeViewItem_MouseDoubleClick(object sender, MouseButtonEventArgs e)
        {
            TreeViewItem? item = GetParentObjectEx<TreeViewItem>(e.OriginalSource as DependencyObject) as TreeViewItem;
            if (item == null) { return; }

            item.Focus();
            e.Handled = true;

            tempTextBlock = FindVisualChild<TextBlock>(item as DependencyObject);
            if (tempTextBlock != null)
            {
                tempTextBlock.Visibility = Visibility.Collapsed;
            }

            TextBox? tempTextBox = FindVisualChild<TextBox>(item as DependencyObject);
            if (tempTextBox != null) 
            {
                tempTextBox.Visibility = Visibility.Visible;
                tempTextBox.Focus();
                tempTextBox.Select(tempTextBox.Text.Length, 0);
            }
        }

        private void renametextbox_LostFocus(object sender, RoutedEventArgs e)
        {
            TextBox tempTextBox = sender as TextBox;
            if (tempTextBox != null) { 
                tempTextBox.Visibility = Visibility.Collapsed; 
            }
            if (tempTextBlock != null)
            {
                tempTextBlock.Visibility = Visibility.Visible;
                tempTextBlock = null;
            }
        }

        private void renametextbox_KeyDown(object sender, KeyEventArgs e)
        {
            if (e.Key == Key.Enter)
            {
                TextBox tempTextBox = sender as TextBox;
                if (tempTextBox != null)
                {
                    tempTextBox.MoveFocus(new TraversalRequest(FocusNavigationDirection.Previous));
                    e.Handled = true;
                }
            }
        }

        private TreeViewItem? GetParentObjectEx<TreeViewItem>(DependencyObject obj) where TreeViewItem : FrameworkElement
        {
            DependencyObject parent = VisualTreeHelper.GetParent(obj);
            while (parent != null)
            {
                if (parent is TreeViewItem)
                {
                    return (TreeViewItem)parent;
                }
                parent = VisualTreeHelper.GetParent(parent);
            }
            return null;
        }

        private childItem? FindVisualChild<childItem>(DependencyObject? obj) where childItem : DependencyObject
        {
            for (int i = 0; i < VisualTreeHelper.GetChildrenCount(obj); i++)
            {
                DependencyObject child = VisualTreeHelper.GetChild(obj, i);
                if (child != null && child is childItem)
                    return (childItem)child;
                else
                {
                    childItem? childOfChild = FindVisualChild<childItem>(child);
                    if (childOfChild != null)
                        return childOfChild;
                }
            }
            return null;
        }


    }
}
