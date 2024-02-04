using Behavior.Diagrams.Controls;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Behavior.Diagrams
{
    public interface IDiagramController
    {
        /// <summary>
        /// 当用户移动/调整项目大小时调用
        /// </summary>
        /// <param name="items">所选项目</param>
        /// <param name="bounds">新项目范围</param>
        void UpdateItemsBounds(DiagramItem[] items, Rect[] bounds);
        /// <summary>
        /// 当用户在项目之间创建链接时调用
        /// </summary>
        /// <param name="initialState">用户操作之前链接的状态</param>
        /// <param name="link">当前状态下的链接</param>
        void UpdateLink(LinkInfo initialState, ILink link);
    }
}
