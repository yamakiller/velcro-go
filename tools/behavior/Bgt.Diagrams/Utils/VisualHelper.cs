
using System.Windows.Media;
using System.Windows;

namespace Bgt.Diagrams.Utils
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

        /*public static Point GetWindowPosition(this System.Windows.Input.MouseEventArgs e, DependencyObject relativeTo)
		{
			var parentWindow = Window.GetWindow(relativeTo);
			return e.GetPosition(parentWindow);
		}*/

        /*public static Point ClientToScreen(this UIElement value, Point point)
		{
			var parentWindow = Window.GetWindow(value);
			return value.TranslatePoint(point, parentWindow);
		}*/
    }
}
