


namespace Bga.Diagrams.Controls
{
    public interface INode
    {
        IEnumerable<IPort> Ports { get; }
    }
}
