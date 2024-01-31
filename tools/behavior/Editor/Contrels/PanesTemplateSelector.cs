
using AvalonDock.Layout;
using Editor.BehaviorCharts.Model;
using System.Windows;
using System.Windows.Controls;

namespace Editor.Contrels
{

    class PanesTemplateSelector : DataTemplateSelector
    {
        public PanesTemplateSelector()
        {

        }

        public DataTemplate BehaviorChartEditorViewTemplate
        {
            get;
            set;
        }
        public override DataTemplate SelectTemplate(object item, DependencyObject container)
        {
            var itemAsLayoutContent = item as LayoutContent;
            if (item is EditorViewModel)
            {
                return BehaviorChartEditorViewTemplate;
            }
                
            return base.SelectTemplate(item, container);
        }
    }
}
