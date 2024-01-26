using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;

namespace Editor.Contrels
{
    class PanesStyleSelector : StyleSelector
    {
        public Style EditorBehaviorStyle { get; set; }

        public override System.Windows.Style SelectStyle(object item, System.Windows.DependencyObject container)
        {
            if (item is EditorBehaviorViewModel)
                return EditorBehaviorStyle;

            return base.SelectStyle(item, container);
        }
    }
}
