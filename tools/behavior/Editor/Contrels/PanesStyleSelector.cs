
using Editor.Panels.Model;
using System.Windows;
using System.Windows.Controls;

namespace Editor.Contrels
{
    class PanesStyleSelector : StyleSelector
    {
        public Style EditorBehaviorStyle { get; set; }

        public override System.Windows.Style SelectStyle(object item, DependencyObject container)
        {
            if (item is PanelViewModel)
                return EditorBehaviorStyle;

            return base.SelectStyle(item, container);
        }
    }
}
