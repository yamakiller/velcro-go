using System.Windows.Media;
using System.Windows;

namespace Behavior.Diagrams.Utils
{
    public static class VisualHelper
    {
        public static T FindParent<T>(this DependencyObject value) where T : DependencyObject
        {
            DependencyObject parent = value;
            while (parent != null && !(parent is T))
                parent = VisualTreeHelper.GetParent(parent);
            return parent as T;
        }
    }
}
