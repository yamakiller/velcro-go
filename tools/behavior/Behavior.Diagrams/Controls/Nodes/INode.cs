
using System.Collections.Generic;

namespace Behavior.Diagrams.Controls
{
    public interface INode
    {
        IEnumerable<IPort> Ports { get; }
    }
}
