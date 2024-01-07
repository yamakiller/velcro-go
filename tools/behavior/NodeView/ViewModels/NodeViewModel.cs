using ReactiveUI;
using Splat;
using System;
using System.Collections.Generic;
using System.Drawing;
using System.Linq;
using System.Reactive.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NodeView.ViewModels
{

    public enum ResizeOrientation
    {
        None,
        Horizontal,
        Vertical,
        HorizontalAndVertical
    }

    public class NodeViewModel : ReactiveObject
    {
        static NodeViewModel()
        {
            NViewRegistrar.AddRegistration(() => new Views.NodeView(), typeof(IViewFor<NodeViewModel>));
            Locator.CurrentMutable.RegisterPlatformBitmapLoader();
        }

        #region Parent
        /// <summary>
        /// The network that contains this node
        /// </summary>
        public BehaviorViewModel Parent
        {
            get => m_parent;
            internal set => this.RaiseAndSetIfChanged(ref m_parent, value);
        }
        private BehaviorViewModel m_parent;
        #endregion

        #region Name
        /// <summary>
        /// The name of the node.
        /// In the default view, this string is displayed at the top of the node.
        /// </summary>
        public string Name
        {
            get => m_name;
            set => this.RaiseAndSetIfChanged(ref m_name, value);
        }
        private string m_name;
        #endregion

        #region HeaderIcon
        /// <summary>
        /// The icon displayed in the header of the node.
        /// If this is null, no icon is displayed.
        /// In the default view, this icon is displayed at the top of the node.
        /// </summary>
        public IBitmap HeaderIcon
        {
            get => m_headerIcon;
            set => this.RaiseAndSetIfChanged(ref m_headerIcon, value);
        }
        private IBitmap m_headerIcon;
        #endregion

        #region IsSelected
        /// <summary>
        /// 如果是 true, 节点已被选中.
        /// </summary>
        public bool IsSelected
        {
            get => m_isSelected;
            set => this.RaiseAndSetIfChanged(ref m_isSelected, value);
        }
        private bool m_isSelected;
        #endregion

        #region IsCollapsed
        /// <summary>
        /// 如果是 true, 节点已折叠.
        /// </summary>
        public bool IsCollapsed
        {
            get => _isCollapsed;
            set => this.RaiseAndSetIfChanged(ref _isCollapsed, value);
        }
        private bool _isCollapsed;
        #endregion

        #region CanBeRemovedByUser
        /// <summary>
        /// 如果为 true，则用户可以在 UI 中从视图中删除此节点.默认为真.
        /// </summary>
        public bool CanBeRemovedByUser
        {
            get => _canBeRemovedByUser;
            set => this.RaiseAndSetIfChanged(ref _canBeRemovedByUser, value);
        }
        private bool _canBeRemovedByUser;
        #endregion

        #region Position
        /// <summary>
        /// 节点在视图中的位置
        /// </summary>
        public Point Position
        {
            get => m_position;
            set => this.RaiseAndSetIfChanged(ref m_position, value);
        }
        private Point m_position;
        #endregion

        #region Size
        /// <summary>
        /// The rendered size of this node.
        /// </summary>
        public Size Size
        {
            get => _size;
            internal set => this.RaiseAndSetIfChanged(ref _size, value);
        }
        private Size _size;
        #endregion

        #region Resizable
        /// <summary>
        /// On which axes can the user resize the node?
        /// </summary>
        public ResizeOrientation Resizable
        {
            get => _resizable;
            set => this.RaiseAndSetIfChanged(ref _resizable, value);
        }
        private ResizeOrientation _resizable;
        #endregion

        public NodeViewModel()
        {
            this.Name = "Untitled";
            this.CanBeRemovedByUser = true;
            this.Resizable = ResizeOrientation.Horizontal;

            // If collapsed, hide inputs without connections, otherwise show all.
            var onCollapseChange = this.WhenAnyValue(vm => vm.IsCollapsed).Publish();
            onCollapseChange.Connect();
        }
    }
}
