using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Documents;
using System.Windows.Media;

namespace NodeBehavior.Views
{
    public class NodeBehaviorView : Canvas
    {
        private Pen m_gridPen; // 网格线
        private Adorner m_dragAdorner;

        public Adorner DragAdorner
        {
            get { return m_dragAdorner; }
            set
            {
                if (m_dragAdorner != value)
                {
                    var adornerLayer = AdornerLayer.GetAdornerLayer(this);
                    if (m_dragAdorner != null)
                        adornerLayer.Remove(m_dragAdorner);
                    m_dragAdorner = value;
                    if (m_dragAdorner != null)
                        adornerLayer.Add(m_dragAdorner);
                }
            }
        }

        static NodeBehaviorView()
        {
            FrameworkElement.DefaultStyleKeyProperty.OverrideMetadata(
                typeof(NodeBehaviorView), new FrameworkPropertyMetadata(typeof(NodeBehaviorView)));
        }

        #region 属性
        #region 网格大小 
        public static readonly DependencyProperty GridCellSizeProperty =
            DependencyProperty.Register("GridCellSize",
                                       typeof(Size),
                                       typeof(NodeBehaviorView),
                                       new FrameworkPropertyMetadata(new Size(10, 10)));

        public Size GridCellSize
        {
            get { return (Size)GetValue(GridCellSizeProperty); }
            set { SetValue(GridCellSizeProperty, value); }
        }
        #endregion

        #region 显示网格
        public static readonly DependencyProperty ShowGridProperty =
            DependencyProperty.Register("ShowGrid",
                                       typeof(bool),
                                       typeof(NodeBehaviorView),
                                       new FrameworkPropertyMetadata(false));

        public bool ShowGrid
        {
            get { return (bool)GetValue(ShowGridProperty); }
            set { SetValue(ShowGridProperty, value); }
        }
        #endregion

        #region 文档网格

        public static readonly DependencyProperty DocumentSizeProperty =
            DependencyProperty.Register("DocumentSize",
                                       typeof(Size),
                                       typeof(NodeBehaviorView),
                                       new FrameworkPropertyMetadata(new Size(2000, 2000)));

        public Size DocumentSize
        {
            get { return (Size)GetValue(DocumentSizeProperty); }
            set { SetValue(DocumentSizeProperty, value); }
        }

        #endregion

        #region Zoom

        public static readonly DependencyProperty ZoomProperty =
            DependencyProperty.Register("Zoom",
                                       typeof(double),
                                       typeof(NodeBehaviorView),
                                       new FrameworkPropertyMetadata(1.0, new PropertyChangedCallback(OnZoomChanged)));

        private static void OnZoomChanged(DependencyObject d, DependencyPropertyChangedEventArgs e)
        {
            var view = d as NodeBehaviorView;
            var zoom = (double)e.NewValue;
            view.m_gridPen = view.CreateGridPen();
            if (Math.Abs(zoom - 1) < 0.0001)
                view.LayoutTransform = null;
            else
                view.LayoutTransform = new ScaleTransform(zoom, zoom);
        }

        public double Zoom
        {
            get { return (double)GetValue(ZoomProperty); }
            set { SetValue(ZoomProperty, value); }
        }

        #endregion


        public bool IsDragging { get { return DragAdorner != null; } }
        #endregion

        public NodeBehaviorView()
        {
            m_gridPen = CreateGridPen();
            Background = new SolidColorBrush(Colors.DarkGray);
            Focusable = true;
        }

        protected override Size MeasureOverride(Size constraint)
        {
            base.MeasureOverride(DocumentSize);
            return DocumentSize;
        }

        protected override void OnRender(DrawingContext dc)
        {
            var rect = new Rect(0, 0, RenderSize.Width, RenderSize.Height);
            dc.DrawRectangle(Background, null, rect);
            if (ShowGrid && GridCellSize.Width > 0 && GridCellSize.Height > 0)
                DrawGrid(dc, rect);
        }


        protected virtual void DrawGrid(DrawingContext dc, Rect rect)
        {
            //using .5 forces wpf to draw a single pixel line
            for (var i = 0.5; i < rect.Height; i += GridCellSize.Height)
                dc.DrawLine(m_gridPen, new Point(0, i), new Point(rect.Width, i));
            for (var i = 0.5; i < rect.Width; i += GridCellSize.Width)
                dc.DrawLine(m_gridPen, new Point(i, 0), new Point(i, rect.Height));
        }

        protected virtual Pen CreateGridPen()
        {
            return new Pen(Brushes.LightGray, (1 / Zoom));
        }
    }
}
