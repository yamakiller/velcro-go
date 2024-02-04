using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Documents;
using System.Windows.Media;

namespace Behavior.Diagrams.Adorners
{
    public class SelectionAdorner : Adorner
    {
        #region 成员
        private VisualCollection m_visuals; // 绘图器接口
        private Control m_control; // 控件
        #endregion

        /// <summary>
        /// 子对象数量
        /// </summary>
        protected override int VisualChildrenCount
        {
            get { return m_visuals.Count; }
        }

        /// <summary>
        /// 构造
        /// </summary>
        /// <param name="item"></param>
        /// <param name="control"></param>
        public SelectionAdorner(Controls.DiagramItem item, Control control)
         : base(item)
        {
            this.m_control = control;
            control.DataContext = item;
            m_visuals = new VisualCollection(this);
            m_visuals.Add(control);
        }

        protected override Size ArrangeOverride(Size finalSize)
        {
            m_control.Arrange(new Rect(finalSize));
            return finalSize;
        }

        protected override Visual GetVisualChild(int index)
        {
            return m_visuals[index];
        }
    }
}
