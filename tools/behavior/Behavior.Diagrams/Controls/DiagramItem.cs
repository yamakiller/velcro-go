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
    public abstract class DiagramItem : Control
    {
        #region 属性

        private Adorner selectionAdorner { get; set; }

        /// <summary>
        /// Model 元素
        /// </summary>
        public object ModelElement { get; set; }

        #region 是否已被选择
        internal static readonly DependencyProperty IsSelectedProperty =
          DependencyProperty.Register("IsSelected",
                                     typeof(bool),
                                     typeof(DiagramItem),
                                     new FrameworkPropertyMetadata(false, new PropertyChangedCallback(onIsSelectedChanged)));

        public bool IsSelected
        {
            get { return (bool)GetValue(IsSelectedProperty); }
            internal set { SetValue(IsSelectedProperty, value); }
        }

        private static void onIsSelectedChanged(DependencyObject d, DependencyPropertyChangedEventArgs e)
        {
            if (!(bool)e.NewValue)
            {
                d.ClearValue(IsPrimarySelectionProperty);
                (d as DiagramItem).hideSelectionAdorner();
            }
            else
            {
                (d as DiagramItem).showSelectionAdorner();
            }
        }
        #endregion

        #region IsPrimarySelection 属性
        internal static readonly DependencyProperty IsPrimarySelectionProperty =
           DependencyProperty.Register("IsPrimarySelection",
                                      typeof(bool),
                                      typeof(DiagramItem),
                                      new FrameworkPropertyMetadata(false));

        public bool IsPrimarySelection
        {
            get { return (bool)GetValue(IsPrimarySelectionProperty); }
            internal set { SetValue(IsPrimarySelectionProperty, value); }
        }
        #endregion

        #region 是否支持移动
        public bool IsMove
        {
            get { return (bool)GetValue(IsMoveProperty); }
            set { SetValue(IsMoveProperty, value); }
        }

        private static readonly DependencyProperty IsMoveProperty =
           DependencyProperty.Register("IsMove",
                                      typeof(bool),
                                      typeof(DiagramItem),
                                      new FrameworkPropertyMetadata(true));
        #endregion

        #region 是否支持选择

        public bool IsSelect
        {
            get { return (bool)GetValue(IsSelectProperty); }
            set { SetValue(IsSelectProperty, value); }
        }

        private static readonly DependencyProperty IsSelectProperty =
          DependencyProperty.Register("IsSelect",
                                     typeof(bool),
                                     typeof(DiagramItem),
                                     new FrameworkPropertyMetadata(true));
        #endregion

        public abstract Rect Bounds { get; }

        #endregion

        protected void hideSelectionAdorner()
        {
            if (selectionAdorner != null)
            {
                AdornerLayer adornerLayer = AdornerLayer.GetAdornerLayer(this);
                if (adornerLayer != null)
                    adornerLayer.Remove(selectionAdorner);
                selectionAdorner = null;
            }
        }

        protected void showSelectionAdorner()
        {
            var adornerLayer = AdornerLayer.GetAdornerLayer(this);
            if (adornerLayer != null)
            {
                selectionAdorner = CreateSelectionAdorner();
                selectionAdorner.Visibility = Visibility.Visible;
                adornerLayer.Add(selectionAdorner);
            }
        }

        protected abstract Adorner CreateSelectionAdorner();
    }
}
