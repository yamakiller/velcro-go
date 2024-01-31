
using System.Windows.Input;

namespace Bgt.Diagrams.Tools
{
    public interface IInputTool
    {
        void OnMouseDown(MouseButtonEventArgs e);
        void OnMouseMove(MouseEventArgs e);
        void OnMouseUp(MouseButtonEventArgs e);
        void OnPreviewKeyDown(KeyEventArgs e);
    }
}
