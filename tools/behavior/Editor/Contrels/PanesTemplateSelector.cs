
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

        public DataTemplate BehaviorChartViewTemplate
        {
            get;
            set;
        }
        public override DataTemplate SelectTemplate(object item, DependencyObject container)
        {
            var itemAsLayoutContent = item as LayoutContent;
            if (item is BehaviorChartModel)
                return BehaviorChartViewTemplate;
            return base.SelectTemplate(item, container);
        }
    }
}
