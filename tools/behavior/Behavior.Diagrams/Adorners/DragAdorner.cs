
using System.Windows;
using System.Windows.Documents;
using System.Windows.Input;

namespace Behavior.Diagrams.Adorners
{
    public abstract class DragAdorner : Adorner
    {
        #region 属性
        public DiagramView View { get; private set; }
        protected bool DoCommit { get; set; }
        /// <summary>
        /// 是否支持拖拽
        /// </summary>
        private bool IsDrop {  get; set; }
        /// <summary>
        /// 开始位置
        /// </summary>
        protected Point Start { get; set; }
        /// <summary>
        /// 结束位置
        /// </summary>
        protected Point End { get; set; }
        #endregion

        protected DragAdorner(DiagramView view, Point start) : base(view)
        {
            View = view;
            End = Start = start;
            this.Loaded += OnLoaded;
        }

        private void OnLoaded(object sender, RoutedEventArgs e)
        {
            DoCommit = false;
            CaptureMouse();
        }

        protected override void OnMouseMove(System.Windows.Input.MouseEventArgs e)
        {
            End = e.GetPosition(View);
            IsDrop = DoDrag();
            Mouse.OverrideCursor = IsDrop ? Cursor : Cursors.No;
        }

        protected override void OnMouseUp(System.Windows.Input.MouseButtonEventArgs e)
        {
            if (this.IsMouseCaptured)
            {
                DoCommit = IsDrop;
                this.ReleaseMouseCapture();
            }
        }

        protected override void OnLostMouseCapture(MouseEventArgs e)
        {
            View.DragAdorner = null;
            Mouse.OverrideCursor = null;
            EndDrag();
        }

        protected abstract bool DoDrag();
        protected abstract void EndDrag();
    }
}
