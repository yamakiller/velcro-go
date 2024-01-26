
using AvalonDock.Layout;
using Editor.ViewModels;
using System.Windows;
using System.Windows.Controls;

namespace Editor.Contrels
{

    class PanesTemplateSelector : DataTemplateSelector
    {
        public PanesTemplateSelector()
        {

        }

        public DataTemplate EditorBehaviorViewTemplate
        {
            get;
            set;
        }
        public override DataTemplate SelectTemplate(object item, DependencyObject container)
        {
            var itemAsLayoutContent = item as LayoutContent;
            if (item is EditorBehaviorViewModel)
                return EditorBehaviorViewTemplate;
            return base.SelectTemplate(item, container);
        }
    }
}
