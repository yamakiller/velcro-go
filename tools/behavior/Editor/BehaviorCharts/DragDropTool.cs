﻿using Bga.Diagrams.Tools;
using Bga.Diagrams.Views;
using Editor.BehaviorCharts.Model;
using System;
using System.Collections.Generic;
using System.Data.Common;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Editor.Charts
{
    class DragDropTool : IDragDropTool
    {
        DiagramView m_view;
        EditorViewModel m_model;
        int m_row, m_column;

        public DragDropTool(DiagramView view, EditorViewModel model)
        {
            m_view = view;
            m_model = model;
        }

        public void OnDragEnter(System.Windows.DragEventArgs e)
        {
        }

        public void OnDragOver(System.Windows.DragEventArgs e)
        {
            e.Effects = DragDropEffects.None;
            if (e.Data.GetDataPresent(typeof(NodeKinds)))
            {
                var position = e.GetPosition(m_view);
                m_column = (int)(position.X / m_view.GridCellSize.Width);
                m_row = (int)(position.Y / m_view.GridCellSize.Height);
                if (m_column >= 0 && m_row >= 0)
                    if (m_model.Nodes.Where(p => p.Column == m_column && p.Row == m_row).Count() == 0)
                        e.Effects = e.AllowedEffects;
            }
            e.Handled = true;
        }

        public void OnDragLeave(System.Windows.DragEventArgs e)
        {
        }

        public void OnDrop(System.Windows.DragEventArgs e)
        {
            var node = new BehaviorNode((NodeKinds)e.Data.GetData(typeof(NodeKinds)));
            node.Row = m_row;
            node.Column = m_column;
            m_model.Nodes.Add(node);
            e.Handled = true;
        }
    }
}