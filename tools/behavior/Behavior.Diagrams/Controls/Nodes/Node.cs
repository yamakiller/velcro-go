using Behavior.Diagrams.Adorners;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Documents;

namespace Behavior.Diagrams.Controls
{
    public class Node : DiagramItem, INode
    {
        static Node()
        {
            FrameworkElement.DefaultStyleKeyProperty.OverrideMetadata(
                typeof(Node), new FrameworkPropertyMetadata(typeof(Node)));
        }

        #region 属性

        #region 内容属性
        public object Content
        {
            get { return (bool)GetValue(ContentProperty); }
            set { SetValue(ContentProperty, value); }
        }

        public static readonly DependencyProperty ContentProperty =
            DependencyProperty.Register("Content",
                                       typeof(object),
                                       typeof(Node));
        #endregion


        #region 是否支持大小更改属性
        public bool IsResize
        {
            get { return (bool)GetValue(IsResizeProperty); }
            set { SetValue(IsResizeProperty, value); }
        }

        public static readonly DependencyProperty IsResizeProperty =
            DependencyProperty.Register("CanResize",
                                       typeof(bool),
                                       typeof(Node),
                                       new FrameworkPropertyMetadata(true));
        #endregion

        #region 可连接的节点
        private List<IPort> m_ports = new List<IPort>();
        public ICollection<IPort> Ports { get { return m_ports; } }

        IEnumerable<IPort> INode.Ports
        {
            get { return m_ports; }
        }

        #endregion


        public override Rect Bounds
        {
            get
            {
                //var itemRect = VisualTreeHelper.GetDescendantBounds(item);
                //return item.TransformToAncestor(this).TransformBounds(itemRect);
                var x = Canvas.GetLeft(this);
                var y = Canvas.GetTop(this);
                return new Rect(x, y, ActualWidth, ActualHeight);
            }
        }

        #endregion

        public Node()
        {
        }

        public void UpdatePosition()
        {
            foreach (var p in Ports)
                p.UpdatePosition();
        }

        protected override Adorner CreateSelectionAdorner()
        {
            return new SelectionAdorner(this, new SelectionFrame());
        }
    }
}
