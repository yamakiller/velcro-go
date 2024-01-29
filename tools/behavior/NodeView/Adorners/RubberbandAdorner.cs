
using System.Windows.Media;
using System.Windows;

using Bga.Diagrams.Views;

namespace Bga.Diagrams.Adorners
{
    class RubberbandAdorner : DragAdorner
    {
        private Pen m_pen;

        public RubberbandAdorner(DiagramView view, Point start)
            : base(view, start)
        {
            m_pen = new Pen(Brushes.Black, 2);
        }

        protected override bool DoDrag()
        {
            InvalidateVisual();
            return true;
        }

        protected override void EndDrag()
        {
            if (DoCommit)
            {
                var rect = new Rect(Start, End);
                var items = View.Items.Where(p => p.CanSelect && rect.Contains(p.Bounds));
                View.Selection.SetRange(items);
            }
        }

        protected override void OnRender(DrawingContext dc)
        {
            dc.DrawRectangle(Brushes.Transparent, m_pen, new Rect(Start, End));
        }
    }
}
