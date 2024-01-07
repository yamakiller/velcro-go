using DynamicData;
using NodeView.ViewModels;
using ReactiveUI;
using System.Reactive.Linq;
using System.Windows;


namespace NodeBehavior.ViewModels
{
    public class SelectionRectangleViewModel : ReactiveObject
    {
        #region StartPoint
        /// <summary>
        /// The coordinates of the first corner of the rectangle (where the user clicked down).
        /// </summary>
        public Point StartPoint
        {
            get => m_startPoint;
            set => this.RaiseAndSetIfChanged(ref m_startPoint, value);
        }
        private Point m_startPoint;
        #endregion

        #region EndPoint
        /// <summary>
        /// The coordinates of the second corner of the rectangle.
        /// </summary>
        public Point EndPoint
        {
            get => m_endPoint;
            set => this.RaiseAndSetIfChanged(ref m_endPoint, value);
        }
        private Point m_endPoint;
        #endregion

        #region Rectangle
        /// <summary>
        /// The Rect object formed by StartPoint and EndPoint.
        /// </summary>
        public Rect Rectangle => m_rectangle.Value;
        private readonly ObservableAsPropertyHelper<Rect> m_rectangle;
        #endregion

        #region IsVisible
        /// <summary>
        /// If true, the selection rectangle view is visible.
        /// </summary>
        public bool IsVisible
        {
            get => m_isVisible;
            set => this.RaiseAndSetIfChanged(ref m_isVisible, value);
        }
        private bool m_isVisible;
        #endregion

        #region IntersectingNodes
        /// <summary>
        /// List of nodes visually intersecting or contained in the rectangle.
        /// This list is driven by the view.
        /// </summary>
        public ISourceList<NodeViewModel> IntersectingNodes { get; } = new SourceList<NodeViewModel>();
        #endregion

        public SelectionRectangleViewModel()
        {
            this.WhenAnyValue(vm => vm.StartPoint, vm => vm.EndPoint)
                .Select(_ => new Rect(StartPoint, EndPoint))
                .ToProperty(this, vm => vm.Rectangle, out m_rectangle);

            IntersectingNodes.Connect().ActOnEveryObject(node => node.IsSelected = true, node => node.IsSelected = false);
        }
    }
}
