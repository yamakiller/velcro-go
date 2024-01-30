
using Editor.Datas;
using Editor.Framework;
using Editor.ViewModels;
using System.Collections.ObjectModel;
using System.Windows.Input;


namespace Editor.BehaviorCharts.Model
{
    class BehaviorChartModel : PaneViewModel
    {
        private EditorFrameViewModel m_parentViewModel;
        private BehaviorTree m_btree;
        private PaneCommand  m_closeCommand;

        #region 字段
        public string FileName
        {
            get { return m_btree.FileName; }
        }

        public ICommand CloseCommand
        {
            get
            {
                if (m_closeCommand == null)
                {
                    m_closeCommand = new PaneCommand((p) => OnClose(), (p) => CanClose());
                }
                return m_closeCommand;
            }
        }
        #endregion

        #region methods
        private bool CanClose()
        {
            return true;
        }

        private void OnClose()
        {
            m_parentViewModel.CloseBehaviorTreeView(m_btree);
        }

        private bool CanSave(object parameter)
        {
            return true;
        }

        private void OnSave(object parameter)
        {

        }

        private bool CanSaveAs(object parameter)
        {
            return true;
        }

        private void OnSaveAs(object parameter)
        {

        }
        #endregion

        private ObservableCollection<BehaviorNode> m_nodes = new ObservableCollection<BehaviorNode> ();
        internal ObservableCollection<BehaviorNode> Nodes
        { 
            get { return m_nodes; } 
        }

        private ObservableCollection<Link> m_links = new ObservableCollection<Link> ();
        internal ObservableCollection<Link> Links
        {
            get { return this.m_links; }
        }

        public BehaviorChartModel(EditorFrameViewModel viewModel, BehaviorTree t)
        {
            m_parentViewModel = viewModel;
            m_btree = t;

            ContentId = t.ID;
            Title = t.Title;
        }
    }
}
