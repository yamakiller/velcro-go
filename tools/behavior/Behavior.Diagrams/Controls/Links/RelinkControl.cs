using System.Windows.Controls;
using System.Windows;

namespace Behavior.Diagrams.Controls
{
    public class RelinkControl : Control
    {
        static RelinkControl()
        {
            FrameworkElement.DefaultStyleKeyProperty.OverrideMetadata(typeof(RelinkControl), new FrameworkPropertyMetadata(typeof(RelinkControl)));
        }
    }
}
