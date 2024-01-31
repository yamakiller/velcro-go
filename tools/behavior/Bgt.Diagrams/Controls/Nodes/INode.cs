


using System.Collections.Generic;

namespace Bgt.Diagrams.Controls
{
    public interface INode
    {
        IEnumerable<IPort> Ports { get; }
    }
}
