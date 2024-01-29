using Bga.Diagrams.Controls.Links;
using Bga.Diagrams.Controls.Ports;
using System.Windows;

namespace Bga.Diagrams.Views
{
    public class LinkInfo
    {
        public IPort Source { get; set; }
        public IPort Target { get; set; }
        public Point? SourcePoint { get; set; }
        public Point? TargetPoint { get; set; }

        public LinkInfo(ILink link)
        {
            Source = link.Source;
            Target = link.Target;
            SourcePoint = link.SourcePoint;
            TargetPoint = link.TargetPoint;
        }

        public void UpdateLink(ILink link)
        {
            link.Source = Source;
            link.Target = Target;
            link.SourcePoint = SourcePoint;
            link.TargetPoint = TargetPoint;
        }
    }
}
