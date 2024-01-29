using Bga.Diagrams.Controls.Links;
using System.Windows;


namespace Bga.Diagrams.Controls.Ports
{
    public interface IPort
    {
        ICollection<ILink> Links { get; }
        Point Center { get; }

        bool IsNear(Point point);
        Point GetEdgePoint(Point target);
        void UpdatePosition();
    }
}
