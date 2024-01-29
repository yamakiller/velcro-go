using Bga.Diagrams.Controls.Links;
using Bga.Diagrams.Controls.Ports;
using System.Windows;

namespace Bga.Diagrams.Tools
{
    public interface ILinkTool
    {
        void BeginDrag(Point start, ILink link, LinkThumbKind thumb);
        void BeginDragNewLink(Point start, IPort port);
        void DragTo(Vector vector);
        bool CanDrop();
        void EndDrag(bool doCommit);
    }
}
