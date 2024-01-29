

using System.Windows;
using System.Windows.Media;

using Bga.Diagrams.Controls.Ports;
using Bga.Diagrams.Views;

namespace Bga.Diagrams.Adorners
{
    public class LinkAdorner : DragAdorner
    {
        private Pen m_pen;

        private IPort m_port;
        public IPort Port
        {
            get { return m_port; }
            set
            {
                if (m_port != value)
                {
                    m_port = value;
                    InvalidateVisual();
                }
            }
        }


        public LinkAdorner(DiagramView view, Point start)
            : base(view, start)
        {
            m_pen = new Pen(new SolidColorBrush(Colors.Red), 1);
        }

        protected override bool DoDrag()
        {
            View.LinkTool.DragTo(End - Start);
            return View.LinkTool.CanDrop();
        }

        protected override void EndDrag()
        {
            View.LinkTool.EndDrag(DoCommit);
        }

        protected override void OnRender(DrawingContext drawingContext)
        {
            if (Port != null)
            {
                var p = Port.Center;
                drawingContext.DrawLine(m_pen, new Point(p.X, p.Y - 3.5), new Point(p.X, p.Y + 3.5));
                drawingContext.DrawLine(m_pen, new Point(p.X - 3.5, p.Y), new Point(p.X + 3.5, p.Y));
            }
        }
    }
}
