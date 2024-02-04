using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Behavior.Diagrams.Controls
{
    public interface IPort
    {
        /// <summary>
        /// 节点上的连线
        /// </summary>
        ICollection<ILink> Links { get; }
        /// <summary>
        /// 中心点
        /// </summary>
        Point Center { get; }
        /// <summary>
        /// 是否接近这个点
        /// </summary>
        /// <param name="point"></param>
        /// <returns></returns>
        bool IsNear(Point point);
        /// <summary>
        /// 边缘点(当前的目标点)
        /// </summary>
        /// <param name="target"></param>
        /// <returns></returns>
        Point GetEdgePoint(Point target);

        /// <summary>
        /// 更新位置
        /// </summary>
        void UpdatePosition();
    }
}
