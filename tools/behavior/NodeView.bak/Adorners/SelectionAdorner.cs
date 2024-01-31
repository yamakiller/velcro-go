

using System.Windows;
using System.Windows.Controls;
using System.Windows.Documents;
using System.Windows.Media;

using Bga.Diagrams.Controls;

namespace Bga.Diagrams.Adorners
{
    public class SelectionAdorner : Adorner
    {
        private VisualCollection m_visuals;
        private Control m_control;

        protected override int VisualChildrenCount
        {
            get { return m_visuals.Count; }
        }

        public SelectionAdorner(DiagramItem item, Control control)
            : base(item)
        {
            this.m_control = control;
            control.DataContext = item;
            m_visuals = new VisualCollection(this);
            m_visuals.Add(control);
        }

        protected override Size ArrangeOverride(Size finalSize)
        {
            m_control.Arrange(new Rect(finalSize));
            return finalSize;
        }

        protected override Visual GetVisualChild(int index)
        {
            return m_visuals[index];
        }
    }
}
