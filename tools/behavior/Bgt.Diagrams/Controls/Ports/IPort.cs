
using System.Collections.Generic;
using System.Windows;


namespace Bgt.Diagrams.Controls
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
