
using Bga.Diagrams.Controls;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Bga.Diagrams.Tools
{
    public interface IMoveResizeTool
    {
        void BeginDrag(Point start, DiagramItem item, DragThumbKinds kind);
        void DragTo(Vector vector);
        bool CanDrop();
        void EndDrag(bool doCommit);
    }
}
