
using Bga.Diagrams.Controls.Ports;
using System.Windows;

namespace Bga.Diagrams.Controls.Links
{
    public interface ILink
    {
        IPort Source { get; set; }
        IPort Target { get; set; }
        Point? SourcePoint { get; set; }
        Point? TargetPoint { get; set; }

        void UpdatePath();
    }
}
