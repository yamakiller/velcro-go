
using Bgt.Diagrams;
using Editor.Charts;
using Editor.Datas;
using Editor.Framework;
using Editor.ViewModels;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Windows.Input;


namespace Editor.BehaviorCharts.Model
{
    class EditorViewModel : PaneViewModel
    {
        private EditorFrameViewModel m_parentViewModel;
        private BehaviorTree m_btree;
        

        #region 字段
        public string FileName
        {
            get { return m_btree.FileName; }
        }

        private PaneCommand m_loadedCommand;
        public ICommand LoadedCommand
        {
            get
            {
                if (m_loadedCommand == null)
                {
                    m_loadedCommand = new PaneCommand((p) => OnLoaded(p), (p) => CanLoaded(p));
                }
                return m_loadedCommand;
            }
        }

        private PaneCommand m_closeCommand;
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

        private bool CanLoaded(object parameter)
        {
            return true;
        }

        private void OnLoaded(object parameter)
        {
            DiagramView editor = parameter as DiagramView;
            Debug.Assert(editor != null);

            var start = new BehaviorNode(NodeKinds.Root);
            start.Row = 2;
            start.Column = 2;
            start.Text = "Start";

            this.Nodes.Add(start);

            editor.Controller = new Controller(editor, this);
            editor.DragDropTool = new DragDropTool(editor, this);
            editor.DragTool = new CustomMoveResizeTool(editor, this)
            {
                //MoveGridCell = editor.GridCellSize
            };

            editor.Selection.PropertyChanged += new System.ComponentModel.PropertyChangedEventHandler(Selection_PropertyChanged);
        }

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

        public EditorViewModel(EditorFrameViewModel viewModel, BehaviorTree t)
        {
            m_parentViewModel = viewModel;
            m_btree = t;

            ContentId = t.ID;
            Title = t.Title;
        }

        void Selection_PropertyChanged(object sender, System.ComponentModel.PropertyChangedEventArgs e)
        {
            int a = 100;
            if (a == 100)
            {

            }
            // var p = editor.Selection.Primary;
            // m_propertiesView.SelectedObject = p != null ? p.ModelElement : null;
        }
    }
}
