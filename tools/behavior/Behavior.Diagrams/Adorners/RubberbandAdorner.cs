using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Media;

namespace Behavior.Diagrams.Adorners
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
                var items = View.Items.Where(p => p.IsSelect && rect.Contains(p.Bounds));
                View.Selection.SetRange(items);
            }
        }

        protected override void OnRender(DrawingContext dc)
        {
            dc.DrawRectangle(Brushes.Transparent, m_pen, new Rect(Start, End));
        }
    }
}
