using Bga.Diagrams.Controls.Ports;


namespace Bga.Diagrams.Controls
{
    public interface INode
    {
        IEnumerable<IPort> Ports { get; }
    }
}
