
using System.Windows.Controls;
using System.Windows;

namespace Behavior.Diagrams.Controls
{
    public class SelectionFrame : Control
    {
        static SelectionFrame()
        {
            FrameworkElement.DefaultStyleKeyProperty.OverrideMetadata(typeof(SelectionFrame), new FrameworkPropertyMetadata(typeof(SelectionFrame)));
        }
    }
}
