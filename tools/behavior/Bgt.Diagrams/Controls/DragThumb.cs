using Bgt.Diagrams.Utils;

using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;

namespace Bgt.Diagrams.Controls
{
    public class DragThumb : Control
    {
        public DragThumbKinds Kind { get; set; }

        protected Point MouseDownPoint { get; set; }

        protected override void OnMouseDown(System.Windows.Input.MouseButtonEventArgs e)
        {
            if (e.ChangedButton == MouseButton.Left)
            {
                var item = this.DataContext as DiagramItem;
                var view = VisualHelper.FindParent<DiagramView>(item);
                if (item != null && view != null)
                {
                    MouseDownPoint = e.GetPosition(view);
                    view.DragTool.BeginDrag(MouseDownPoint, item, this.Kind);
                    e.Handled = true;
                }
            }
        }
    }
}
