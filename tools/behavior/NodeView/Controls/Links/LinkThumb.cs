﻿using Bga.Diagrams.Utils;
using Bga.Diagrams.Views;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Input;

namespace Bga.Diagrams.Controls
{
    public class LinkThumb : Control
    {
        public LinkThumbKind Kind { get; set; }
        protected Point? MouseDownPoint { get; set; }

        protected override void OnMouseDown(System.Windows.Input.MouseButtonEventArgs e)
        {
            if (e.ChangedButton == MouseButton.Left)
            {
                var link = this.DataContext as LinkBase;
                var view = VisualHelper.FindParent<DiagramView>(link);
                if (link != null && view != null)
                {
                    MouseDownPoint = e.GetPosition(view);
                    view.LinkTool.BeginDrag(MouseDownPoint.Value, link, this.Kind);
                    e.Handled = true;
                }
            }
        }
    }
}