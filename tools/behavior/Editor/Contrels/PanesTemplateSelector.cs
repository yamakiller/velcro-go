
using AvalonDock.Layout;
using Editor.Panels.Model;
using System.Windows;
using System.Windows.Controls;

namespace Editor.Contrels
{

    class PanesTemplateSelector : DataTemplateSelector
    {
        public PanesTemplateSelector()
        {

        }

        public DataTemplate PanelViewTemplate
        {
            get;
            set;
        }

        public override DataTemplate SelectTemplate(object item, DependencyObject container)
        {
            var itemAsLayoutContent = item as LayoutContent;
            if (item is PanelViewModel)
            {
                return PanelViewTemplate;
            }
                
            return base.SelectTemplate(item, container);
        }
    }
}
