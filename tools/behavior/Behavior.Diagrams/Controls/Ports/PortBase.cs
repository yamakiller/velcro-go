using Behavior.Diagrams.Utils;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;

namespace Behavior.Diagrams.Controls
{
    public abstract class PortBase : Control, IPort
    {
        #region 属性
        /// <summary>
        /// 所有连接的子线
        /// </summary>
        private List<ILink> links = new List<ILink>();
        public ICollection<ILink> Links { get { return links; } }

        /// <summary>
        /// 进线
        /// </summary>
        public IEnumerable<ILink> IncomingLinks
        {
            get { return Links.Where(p => p.Target == this); }
        }

        /// <summary>
        /// 出线
        /// </summary>
        public IEnumerable<ILink> OutgoingLinks
        {
            get { return Links.Where(p => p.Source == this); }
        }

        #region 中心点
        /// <summary>
        /// 中心点
        /// </summary>
        private Point m_center;
        public Point Center
        {
            get { return m_center; }
            protected set
            {
                if (m_center != value && !double.IsNaN(value.X) && !double.IsNaN(value.Y))
                {
                    Trace.WriteLine(this);
                    m_center = value;
                    foreach (var link in Links)
                        link.UpdatePath();
                }
            }
        }
        #endregion

        #region 灵敏度

        public double Sensitivity
        {
            get { return (double)GetValue(SensitivityProperty); }
            set { SetValue(SensitivityProperty, value); }
        }

        public static readonly DependencyProperty SensitivityProperty =
            DependencyProperty.Register("Sensitivity",
                                       typeof(double),
                                       typeof(PortBase),
                                       new FrameworkPropertyMetadata(5.0));
        #endregion

        #endregion

        protected PortBase()
        {
        }

        public virtual void UpdatePosition()
        {
            var canvas = VisualHelper.FindParent<Canvas>(this);
            if (canvas != null)
                Center = this.TransformToAncestor(canvas).Transform(new Point(this.ActualWidth / 2, this.ActualHeight / 2));
            else
                Center = new Point(Double.NaN, Double.NaN);
        }

        /// <summary>
        /// 计算端口边界与中心点和目标点之间的线的交点
        /// </summary>
        /// <param name="target"></param>
        /// <returns></returns>
        public abstract Point GetEdgePoint(Point target);
        /// <summary>
        /// 返回指定点是否在端口敏感区域内
        /// </summary>
        /// <param name="point"></param>
        /// <returns></returns>
        public abstract bool IsNear(Point point);

        /// <summary>
        /// 节点鼠标左键按下, 不支持拉线操作
        /// </summary>
        /// <param name="e"></param>
        protected override void OnPreviewMouseLeftButtonDown(System.Windows.Input.MouseButtonEventArgs e)
        {
           base.OnMouseLeftButtonDown(e);
        }
    }
}
