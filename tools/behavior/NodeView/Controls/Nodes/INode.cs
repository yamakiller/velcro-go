using Bga.Diagrams.Controls.Ports;


namespace Bga.Diagrams.Controls.Nodes
{
    public interface INode
    {
        IEnumerable<IPort> Ports { get; }
    }
}
