using Bgt.Diagrams.Tools;
using Bgt.Diagrams;
using Editor.BehaviorCharts.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.BehaviorCharts
{
    class CustomMoveResizeTool : MoveResizeTool
    {
        private EditorViewModel m_model;
        public CustomMoveResizeTool(DiagramView view, EditorViewModel model)
            : base(view)
        {
            m_model = model;
        }

        public override bool CanDrop()
        {
            foreach (var item in DragItems)
            {
                var column = (int)(item.Bounds.X / View.GridCellSize.Width);
                var row = (int)(item.Bounds.Y / View.GridCellSize.Height);
                if (m_model.Nodes.Where(p => !IsDragged(p) && p.Row == row && p.Column == column).Count() != 0)
                    return false;
            }
            return true;
        }

        private bool IsDragged(BehaviorNode node)
        {
            return DragItems.Where(p => p.ModelElement == node).Count() > 0;
        }
    }
}
