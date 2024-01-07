using DynamicData;
using NodeBehavior.ViewModels;
using NodeBehavior.Views;
using ReactiveUI;
using System;
using System.Collections.Generic;
using System.Drawing;
using System.Linq;
using System.Reactive;
using System.Text;
using System.Threading.Tasks;

namespace NodeView.ViewModels
{
    public class BehaviorViewModel : ReactiveObject
    {
        static BehaviorViewModel()
        {
            NViewRegistrar.AddRegistration(() => new BehaviorView(), typeof(IViewFor<BehaviorViewModel>));
        }


        #region Nodes
        /// <summary>
        /// 所有节点
        /// </summary>
        public ISourceList<NodeViewModel> Nodes { get; } = new SourceList<NodeViewModel>();
        #endregion

        #region SelectedNodes
        /// <summary>
        /// A list of nodes that are currently selected in the UI.
        /// The contents of this list is equal to the nodes in Nodes where the Selected property is true.
        /// </summary>
        public IObservableList<NodeViewModel> SelectedNodes { get; }
        #endregion

        #region ZoomFactor
        /// <summary>
        /// Scale of the view. Larger means more zoomed in. Default value is 1.
        /// </summary>
        public double ZoomFactor
        {
            get => m_zoomFactor;
            set => this.RaiseAndSetIfChanged(ref m_zoomFactor, value);
        }

        private double m_zoomFactor = 1;

        /// <summary>
        /// The maximum zoom level used in this network view. Default value is 2.5.
        /// </summary>
        public double MaxZoomLevel
        {
            get => m_maxZoomLevel;
            set => this.RaiseAndSetIfChanged(ref m_maxZoomLevel, value);
        }

        private double m_maxZoomLevel = 2.5;

        /// <summary>
        /// The minimum zoom level used in this network view. Default value is 0.15.
        /// </summary>
        public double MinZoomLevel
        {
            get => m_minZoomLevel;
            set => this.RaiseAndSetIfChanged(ref m_minZoomLevel, value);
        }

        private double m_minZoomLevel = 0.15;

        /// <summary>
        /// The drag offset of the initial view position used in this network view. Default value is (0, 0).
        /// </summary>
        public Point DragOffset
        {
            get => m_dragOffset;
            set => this.RaiseAndSetIfChanged(ref m_dragOffset, value);
        }

        private Point m_dragOffset = new Point(0, 0);

        #endregion

        #region SelectionRectangle
        /// <summary>
        /// The viewmodel for the selection rectangle used in this network view.
        /// </summary>
        public SelectionRectangleViewModel SelectionRectangle { get; } = new SelectionRectangleViewModel();
        #endregion

        #region Commands
        /// <summary>
        /// Deletes the nodes in SelectedNodes that are user-removable.
        /// </summary>
        public ReactiveCommand<Unit, Unit> DeleteSelectedNodes { get; }

        /// <summary>
        /// Runs the Validator function and stores the result in LatestValidation.
        /// </summary>
        //public ReactiveCommand<Unit, NetworkValidationResult> UpdateValidation { get; }
        #endregion

        public BehaviorViewModel()
        {
            // Setup parent relationship in nodes.
            Nodes.Connect().ActOnEveryObject(
                addedNode => addedNode.Parent = this,
                removedNode => removedNode.Parent = null
            );

            // SelectedNodes is a derived collection of all nodes with IsSelected = true.
            SelectedNodes = Nodes.Connect()
                .AutoRefresh(node => node.IsSelected)
                .Filter(node => node.IsSelected)
                .AsObservableList();

            // When DeleteSelectedNodes is invoked, remove all nodes that are user-removable and selected.
            DeleteSelectedNodes = ReactiveCommand.Create(() =>
            {
                Nodes.RemoveMany(SelectedNodes.Items.Where(n => n.CanBeRemovedByUser).ToArray());
            });


            //Nodes.Connect().Select((IChangeSet<NodeViewModel> n) => Unit.Default).InvokeCommand(UpdateValidation);
        }

        public void ClearSelection()
        {
            foreach (NodeViewModel node in SelectedNodes.Items)
            {
                node.IsSelected = false;
            }
        }

        /// <summary>
        /// Starts a selection in RectangleSelection
        /// </summary>
        public void StartRectangleSelection()
        {
            ClearSelection();
            SelectionRectangle.IsVisible = true;
            SelectionRectangle.IntersectingNodes.Clear();
        }

        /// <summary>
        /// Stops the current selection in RectangleSelection and applies the changes.
        /// </summary>
        public void FinishRectangleSelection()
        {
            SelectionRectangle.IsVisible = false;
        }
    }
}
