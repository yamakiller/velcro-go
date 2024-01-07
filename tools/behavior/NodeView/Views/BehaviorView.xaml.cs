using NodeView.ViewModels;
using ReactiveUI;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Controls.Primitives;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using System.Windows.Shapes;

namespace NodeBehavior.Views
{
    /// <summary>
    /// BehaviorView.xaml 的交互逻辑
    /// </summary>
    public partial class BehaviorView : IViewFor<BehaviorViewModel>
    {
        #region ViewModel
        public static readonly DependencyProperty ViewModelProperty = DependencyProperty.Register(nameof(ViewModel),
            typeof(BehaviorViewModel), typeof(BehaviorView), new PropertyMetadata(null));

        public BehaviorViewModel ViewModel
        {
            get => (BehaviorViewModel)GetValue(ViewModelProperty);
            set => SetValue(ViewModelProperty, value);
        }

        object IViewFor.ViewModel
        {
            get => ViewModel;
            set => ViewModel = (BehaviorViewModel)value;
        }
        #endregion

        #region BehaviorViewportRegion
        /// <summary>
        /// The rectangle to use as a clipping mask for contentContainer
        /// </summary>
        public Rect BehaviorViewportRegion
        {
            get
            {
                double left = Canvas.GetLeft(contentContainer);
                if (Double.IsNaN(left))
                {
                    left = 0;
                }

                double top = Canvas.GetTop(contentContainer);
                if (Double.IsNaN(top))
                {
                    top = 0;
                }

                if (contentContainer.RenderTransform is ScaleTransform)
                {
                    GeneralTransform transform = this.TransformToDescendant(contentContainer);
                    return transform.TransformBounds(new Rect(0, 0, this.ActualWidth, this.ActualHeight));
                }
                return new Rect(-left, -top, this.ActualWidth, this.ActualHeight);
            }
        }
        private BindingExpressionBase m_viewportBinding;
        #endregion

        #region Node move events
        public class NodeMovementEventArgs : EventArgs
        {
            public IEnumerable<NodeViewModel> Nodes { get; }
            public NodeMovementEventArgs(IEnumerable<NodeViewModel> nodes) => Nodes = nodes.ToList();
        }

        //Start
        public class NodeMoveStartEventArgs : NodeMovementEventArgs
        {
            public DragStartedEventArgs DragEvent { get; }

            public NodeMoveStartEventArgs(IEnumerable<NodeViewModel> nodes, DragStartedEventArgs dragEvent) :
                base(nodes)
            {
                DragEvent = dragEvent;
            }
        }
        public delegate void NodeMoveStartDelegate(object sender, NodeMoveStartEventArgs e);
        /// <summary>Occurs when a (set of) node(s) is selected and starts moving.</summary>
        public event NodeMoveStartDelegate NodeMoveStart;

        //Move
        public class NodeMoveEventArgs : NodeMovementEventArgs
        {
            public DragDeltaEventArgs DragEvent { get; }

            public NodeMoveEventArgs(IEnumerable<NodeViewModel> nodes, DragDeltaEventArgs dragEvent) : base(nodes)
            {
                DragEvent = dragEvent;
            }
        }
        public delegate void NodeMoveDelegate(object sender, NodeMoveEventArgs e);
        /// <summary>Occurs one or more times as the mouse changes position when a (set of) node(s) is selected and has mouse capture.</summary>
        public event NodeMoveDelegate NodeMove;

        //End
        public class NodeMoveEndEventArgs : NodeMovementEventArgs
        {
            public DragCompletedEventArgs DragEvent { get; }

            public NodeMoveEndEventArgs(IEnumerable<NodeViewModel> nodes, DragCompletedEventArgs dragEvent) : base(nodes)
            {
                DragEvent = dragEvent;
            }
        }
        public delegate void NodeMoveEndDelegate(object sender, NodeMoveEndEventArgs e);
        /// <summary>Occurs when a (set of) node(s) loses mouse capture.</summary>
        public event NodeMoveEndDelegate NodeMoveEnd;
        #endregion

        #region BehaviorBackground
        public static readonly DependencyProperty BehaviorBackgroundProperty = DependencyProperty.Register(nameof(BehaviorBackground),
            typeof(Brush), typeof(BehaviorView), new PropertyMetadata(null));

        public Brush BehaviorBackground
        {
            get => (Brush)GetValue(BehaviorBackgroundProperty);
            set => SetValue(BehaviorBackgroundProperty, value);
        }
        #endregion


        /// <summary>
        /// 视图的原点
        /// 可用于计算鼠标相对于视图的位置.
        /// </summary>
        /// <code>
        /// Mouse.GetPosition(network.CanvasOriginElement)
        /// </code>
        public IInputElement CanvasOriginElement => contentContainer;

        #region StartCutGesture
        public static readonly DependencyProperty StartCutGestureProperty = DependencyProperty.Register(nameof(StartCutGesture),
            typeof(MouseGesture), typeof(BehaviorView), new PropertyMetadata(new MouseGesture(MouseAction.RightClick)));

        /// <summary>
        /// This mouse gesture starts a cut, making the cutline visible. Right click by default.
        /// </summary>
        public MouseGesture StartCutGesture
        {
            get => (MouseGesture)GetValue(StartCutGestureProperty);
            set => SetValue(StartCutGestureProperty, value);
        }
        #endregion

        #region StartSelectionRectangleGesture
        public static readonly DependencyProperty StartSelectionRectangleGestureProperty = DependencyProperty.Register(nameof(StartSelectionRectangleGesture),
            typeof(MouseGesture), typeof(BehaviorView), new PropertyMetadata(new MouseGesture(MouseAction.LeftClick, ModifierKeys.Shift)));

        /// <summary>
        /// This mouse gesture starts a selection, making the selection rectangle visible. Left click + Shift by default.
        /// </summary>
        public MouseGesture StartSelectionRectangleGesture
        {
            get => (MouseGesture)GetValue(StartSelectionRectangleGestureProperty);
            set => SetValue(StartSelectionRectangleGestureProperty, value);
        }
        #endregion
        public BehaviorView()
        {
            InitializeComponent();
            if (DesignerProperties.GetIsInDesignMode(this)) { return; }

            //SetupNodes();
            //SetupConnections();
            //SetupCutLine();
            //SetupViewportBinding();
           // SetupKeyboardShortcuts();
           // SetupErrorMessages();
           // SetupDragAndDrop();
           // SetupSelectionRectangle();
        }

        #region Setup
        private void SetupViewportBinding()
        {
            this.WhenActivated(d =>
            {
                this.Bind(ViewModel, vm => vm.ZoomFactor, v => v.dragCanvas.ZoomFactor);
                this.Bind(ViewModel, vm => vm.MaxZoomLevel, v => v.dragCanvas.MaxZoomFactor);
                this.Bind(ViewModel, vm => vm.MinZoomLevel, v => v.dragCanvas.MinZoomFactor);
                this.Bind(ViewModel, vm => vm.DragOffset, v => v.dragCanvas.DragOffset);
            });

            Binding binding = new Binding
            {
                Source = this,
                Path = new PropertyPath(nameof(BehaviorViewportRegion)),
                Mode = BindingMode.OneWay,
                UpdateSourceTrigger = UpdateSourceTrigger.PropertyChanged
            };
            m_viewportBinding = BindingOperations.SetBinding(clippingGeometry, RectangleGeometry.RectProperty, binding);
        }

        private void SetupSelectionRectangle()
        {
            this.WhenActivated(d =>
            {
                this.WhenAnyValue(vm => vm.ViewModel.SelectionRectangle.Rectangle.Left)
                    .Subscribe(left => Canvas.SetLeft(selectionRectangle, left))
                    .DisposeWith(d);
                this.WhenAnyValue(vm => vm.ViewModel.SelectionRectangle.Rectangle.Top)
                    .Subscribe(top => Canvas.SetTop(selectionRectangle, top))
                    .DisposeWith(d);
                this.OneWayBind(ViewModel, vm => vm.SelectionRectangle.Rectangle.Width, v => v.selectionRectangle.Width).DisposeWith(d);
                this.OneWayBind(ViewModel, vm => vm.SelectionRectangle.Rectangle.Height, v => v.selectionRectangle.Height).DisposeWith(d);
                this.OneWayBind(ViewModel, vm => vm.SelectionRectangle.IsVisible, v => v.selectionRectangle.Visibility).DisposeWith(d);

                this.Events().PreviewMouseDown.Subscribe(e =>
                {
                    if (ViewModel != null && StartSelectionRectangleGesture.Matches(this, e))
                    {
                        CaptureMouse();
                        dragCanvas.IsDraggingEnabled = false;
                        ViewModel.StartRectangleSelection();
                        ViewModel.SelectionRectangle.StartPoint = e.GetPosition(contentContainer);
                        ViewModel.SelectionRectangle.EndPoint = ViewModel.SelectionRectangle.StartPoint;
                    }
                }).DisposeWith(d);

                this.Events().MouseMove.Subscribe(e =>
                {
                    if (ViewModel != null && ViewModel.SelectionRectangle.IsVisible)
                    {
                        ViewModel.SelectionRectangle.EndPoint = e.GetPosition(contentContainer);
                        UpdateSelectionRectangleIntersections();
                    }
                }).DisposeWith(d);

                this.Events().MouseUp.Subscribe(e =>
                {
                    if (ViewModel != null && ViewModel.SelectionRectangle.IsVisible)
                    {
                        ViewModel.FinishRectangleSelection();
                        dragCanvas.IsDraggingEnabled = true;
                        ReleaseMouseCapture();
                    }
                }).DisposeWith(d);
            });
        }
        #endregion
    }
}
