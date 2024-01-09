using System.Windows;
using System.Windows.Controls;

namespace Editor.Contrels
{
    public class XToolBar : ToolBar
    {
        public XToolBar() : base() 
        {
            this.Loaded += XToolBar_Loaded;
        }

        private void XToolBar_Loaded(object sender, RoutedEventArgs e)
        {
            ToolBar tbar = sender as ToolBar;
            var overflowGrid = tbar.Template.FindName("OverflowGrid", tbar) as FrameworkElement;
            if (overflowGrid != null )
            {
                overflowGrid.Visibility = Visibility.Collapsed;
            }

            var mainPanelBorder = tbar.Template.FindName("MainPanelBorder", tbar) as FrameworkElement;  
            if (mainPanelBorder != null ) {
                mainPanelBorder.Margin = new Thickness(0);
            }
        }
    }
}
