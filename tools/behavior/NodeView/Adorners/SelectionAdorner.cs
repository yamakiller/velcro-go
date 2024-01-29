

using System.Windows;
using System.Windows.Controls;
using System.Windows.Documents;
using System.Windows.Media;

using Bga.Diagrams.Controls;

namespace Bga.Diagrams.Adorners
{
    public class SelectionAdorner : Adorner
    {
        private VisualCollection visuals;
        private Control control;

        protected override int VisualChildrenCount
        {
            get { return visuals.Count; }
        }

        public SelectionAdorner(DiagramItem item, Control control)
            : base(item)
        {
            this.control = control;
            control.DataContext = item;
            visuals = new VisualCollection(this);
            visuals.Add(control);
        }

        protected override Size ArrangeOverride(Size finalSize)
        {
            control.Arrange(new Rect(finalSize));
            return finalSize;
        }

        protected override Visual GetVisualChild(int index)
        {
            return visuals[index];
        }
    }
}
