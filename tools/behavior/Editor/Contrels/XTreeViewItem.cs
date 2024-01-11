using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;

namespace Editor.Contrels
{
    public class XTreeViewItem : TreeViewItem
    {
        ImageSource iconSource;
        TextBlock textBlock;
        Image icon;

        public XTreeViewItem()
        {
            StackPanel stack = new StackPanel();
            stack.Orientation = Orientation.Horizontal;
            Header = stack;

            icon = new Image();
            icon.VerticalAlignment = VerticalAlignment.Center;
            icon.Height = 10;
            icon.Width = 10;
            icon.Margin = new Thickness(0, 0, 4, 0);
            icon.Source = iconSource;
            stack.Children.Add(icon);
            //Add the HeaderText After Adding the icon
            textBlock = new TextBlock();
            textBlock.VerticalAlignment = VerticalAlignment.Center;
            stack.Children.Add(textBlock);
        }

        public ImageSource Icon
        {
            set
            {
                iconSource = value;
                icon.Source = iconSource;
            }
            get
            {
                return iconSource;
            }
        }

        protected override void OnUnselected(RoutedEventArgs args)
        {
            base.OnUnselected(args);
            icon.Source = iconSource;
        }

        protected override void OnSelected(RoutedEventArgs args)
        {
            base.OnSelected(args);
            icon.Source = iconSource;
        }

        /// <summary>
        /// Gets/Sets the HeaderText of TreeViewWithIcons
        /// </summary>
        public string HeaderText
        {
            set
            {
                textBlock.Text = value;
            }
            get
            {
                return textBlock.Text;
            }
        }
    }
}
