using Editor.Commands;
using Editor.Framework;
using Editor.Panels.Model;
using MaterialDesignThemes.Wpf;
using System.Collections.ObjectModel;
using System.ComponentModel;
using System.Diagnostics;
using System.Drawing;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Media;
using System.Xml;

namespace Editor.ViewModels
{
    class EditorFrameViewModel : ViewModel, ViewModelData
    {
        #region 属性
        // 是否是只读状态
        bool isReadOnly = false;

        public bool IsReadOnly { get { return isReadOnly; } set { SetProperty(ref isReadOnly, value); } }
        // 是否是修改状态
        bool isModifyed = false;

        public bool IsModifyed
        {
            get { return isModifyed; }
            set
            {
                if (value && CurrWorkspace != null)
                {
                    Caption = "*[" + CurrWorkspace.Name + "]" + CurrWorkspace.Dir;
                }
                else if (!value && CurrWorkspace != null)
                {
                    Caption = "[" + CurrWorkspace.Name + "]" + CurrWorkspace.Dir;
                }
                else if (CurrWorkspace == null)
                {
                    Caption = "Behavior Editor";
                }

                SetProperty(ref isModifyed, value);
            }
        }

        private string caption = "Behavior Editor";
        public string Caption
        {
            get { return caption; }
            set { SetProperty(ref caption, value); }
        }
        #endregion

        #region 成员变量
        private Datas.Workspace? wsd = null;
        public Datas.Workspace? CurrWorkspace
        {
            get { return wsd; }
            set
            {
                SetProperty(ref wsd, value);
            }
        }

        private Datas.BehaviorTree? wsdselected = null;
        public Datas.BehaviorTree? CurrWorkspaceSelectedTree
        {
            get { return wsdselected; }
            set
            {
                SetProperty(ref wsdselected, value);
            }
        }

        private bool isWorkspaceExpanded = true;
        public bool IsWorkspaceExpanded
        {
            get { return isWorkspaceExpanded; }
            set
            {


                SetProperty(ref isWorkspaceExpanded, value);
            }
        }

        private bool isWorkspaceDebug = false;
        public bool IsWordspaceDebug
        {
            get { return isWorkspaceDebug; }
            set
            {
                SetProperty(ref isWorkspaceDebug, value);
            }
        }

        #endregion

        #region  Documents
        private ObservableCollection<PanelViewModel> documents;
        public ReadOnlyObservableCollection<PanelViewModel> Documents
        {
            get;
            private set;
        }

        private PanelViewModel activeDocument;
        public PanelViewModel ActiveDocument
        {
            get { return activeDocument; }
            set
            {
                if (activeDocument != value)
                {
                    SetProperty(ref activeDocument, value);
                }
            }
        }

        private object? propertiesSelectedObject = null;
        public object? PropertiesSelectedObject
        {
            get {
                return propertiesSelectedObject;
            }
            set
            {
                var old = propertiesSelectedObject;
                SetProperty(ref propertiesSelectedObject, value);
                if (old != value)
                {
                    DisplayProperties();
                }
            }
        }

        public string propertiesSelectedVisibility;
        public string PropertiesSelectedVisibility
        {
            get { return propertiesSelectedVisibility; }
            set
            {
                string newValue = string.Empty;
                if (propertiesSelectedObject != null)
                {
                    newValue = value;
                }

                SetProperty(ref propertiesSelectedVisibility, newValue);
            }
        }

        private string propertiesSelectedId;
        public string PropertiesSelectedId
        {
            get
            {
                return propertiesSelectedId;
            }

            set
            {
                string newValue = "";
                if (propertiesSelectedObject != null)
                {
                    BNode bnode = propertiesSelectedObject as BNode;
                    if (bnode != null)
                    {
                        var activeNode = activeDocument.FindNode(bnode.Id);
                        if (activeNode != null)
                        {
                            newValue = activeNode.ID;
                        }

                    }
                }

                SetProperty(ref propertiesSelectedId, newValue);
            }
        }

        private string? propertiesSelectedName;
        public string? PropertiesSelectedName
        {
            get
            {
                return propertiesSelectedName;
            }
            set
            {
                string newValue = "";
                if (propertiesSelectedObject != null)
                {
                    BNode bnode = propertiesSelectedObject as BNode;
                    if (bnode != null)
                    {
                        var activeNode = activeDocument.FindNode(bnode.Id);
                        if (activeNode != null)
                        {
                            if (value != null)
                            {
                                newValue = value;
                                activeNode.Name = newValue;
                                bnode.Name = newValue;
                            }
                            else
                            {
                                newValue = activeNode.Name;
                            }
                        }
                    }
                }
                SetProperty(ref propertiesSelectedName, newValue);
            }
        }

        private string? propertiesSelectedCategory;
        public string? PropertiesSelectedCategory
        {
            get {
                return propertiesSelectedCategory;
            }
            set
            {
                string newValue = "";
                if (propertiesSelectedObject != null)
                {
                    BNode bnode = propertiesSelectedObject as BNode;
                    if (bnode != null)
                    {
                        var activeNode = activeDocument.FindNode(bnode.Id);
                        if (activeNode != null)
                        {
                            if (value != null)
                            {
                                var oldCategory = activeNode.Category;
                                newValue = value.ToString();
                                activeNode.Category = newValue;
                                bnode.Category = newValue;
                                if (oldCategory != activeNode.Category)
                                {
                                    PropertiesSelectedColor = NodeKindConvert.ToColor(NodeKindConvert.ToKind(activeNode.Category));
                                }
                            }
                            else
                            {
                                newValue = activeNode.Category;

                            }
                        }
                    }
                }
                SetProperty(ref propertiesSelectedCategory, newValue);
            }
        }

        private string? propertiesSelectedColor;
        public string? PropertiesSelectedColor
        {
            get
            {
                return propertiesSelectedColor;
            }
            set
            {
                string newValue = "";
                if (propertiesSelectedObject != null)
                {
                    BNode bnode = propertiesSelectedObject as BNode;
                    if (bnode != null)
                    {
                        var activeNode = activeDocument.FindNode(bnode.Id);
                        if (activeNode != null)
                        {
                            if (value != null)
                            {
                                newValue = value.ToString();
                                activeNode.Color = newValue;
                                bnode.Color = newValue;
                            }
                            else
                            {
                                newValue = activeNode.Color;
                            }
                        }
                    }
                }

                SetProperty(ref propertiesSelectedColor, newValue);
            }
        }


        private string? propertiesSelectedDesc;
        public string? PropertiesSelectedDesc
        {
            get { return propertiesSelectedDesc; }
            set
            {
                string newValue = "";
                if (propertiesSelectedObject != null)
                {
                    BNode bnode = propertiesSelectedObject as BNode;
                    if (bnode != null)
                    {
                        var activeNode = activeDocument.FindNode(bnode.Id);
                        if (activeNode != null)
                        {
                            if (value != null)
                            {
                                newValue = value;
                                activeNode.Description = newValue;
                                bnode.Description = newValue;
                            }
                            else
                            {
                                newValue = activeNode.Description;
                            }
                        }
                    }
                }
                SetProperty(ref propertiesSelectedDesc, newValue);
            }
        }

        private Dictionary<string, KeyValuePair<string, object>>? propertiesSelectedAttributeMap;
        public Dictionary<string, KeyValuePair<string, object>>? PropertiesSelectedAttributeMap
        {
            get { return propertiesSelectedAttributeMap; }
            set
            {
                Dictionary<string, KeyValuePair<string, object>>? newValue = null;
                if (propertiesSelectedObject != null)
                {
                    BNode bnode = propertiesSelectedObject as BNode;
                    if (bnode != null)
                    {
                        var activeNode = activeDocument.FindNode(bnode.Id);
                        if (activeNode != null)
                        {
                            if (value != null)
                            {
                                newValue = value;
                                activeNode.Properties = newValue;
                            }
                            else
                            {
                                newValue = activeNode.Properties;
                            }
                        }
                    }
                }
                SetProperty(ref propertiesSelectedAttributeMap, newValue);
            }
        }


        void PropertiesPropertyChanged(object sender, PropertyChangedEventArgs e)
        {
            DisplayProperties();
        }

        #endregion

        #region 命令
        private NewWorkspaceCommand? nwcmd = null;
        public NewWorkspaceCommand NewWorkspaceCmd {
            get
            {
                if (nwcmd == null)
                {
                    nwcmd = new NewWorkspaceCommand(this);
                }
                return nwcmd;
            }
        }

        private OpenWorkspaceCommand? opencmd = null;
        public OpenWorkspaceCommand OpenWorkspaceCmd
        {
            get
            {
                if (opencmd == null)
                {
                    opencmd = new OpenWorkspaceCommand(this);
                }
                return opencmd;
            }
        }

        private SaveWorkspaceCommand? savcmd = null;
        public SaveWorkspaceCommand SaveWorkspaceCmd
        {
            get
            {
                if (savcmd == null)
                {
                    savcmd = new SaveWorkspaceCommand(this);
                }
                return savcmd;
            }
        }




        private NewBehaviorTreeCommand? nbtcmd = null;
        public NewBehaviorTreeCommand NewBehaviorTreeCmd
        {
            get
            {
                if (nbtcmd == null)
                {
                    nbtcmd = new NewBehaviorTreeCommand(this);
                }
                return nbtcmd;
            }
        }

        private ExitSystemCommand? excmd = null;

        public ExitSystemCommand ExitSystemCmd
        {
            get
            {
                if (excmd == null)
                {
                    excmd = new ExitSystemCommand(this);
                }
                return excmd;
            }
        }

        private OpenTreeCommand? optcmd = null;
        public OpenTreeCommand OpenTreeCmd
        {
            get
            {
                if (optcmd == null)
                {
                    optcmd = new OpenTreeCommand(this);
                }
                return optcmd;
            }
        }

        private DebugPlayCommand? dpcmd = null;
        public DebugPlayCommand DebugPlayCmd
        {
            get
            {
                if (dpcmd == null)
                {
                    dpcmd = new DebugPlayCommand(this);
                }
                return dpcmd;
            }
        }

        private DebugStopCommand? dscmd = null;
        public DebugStopCommand DebugStopCmd
        {
            get
            {
                if (dscmd == null)
                {
                    dscmd = new DebugStopCommand(this);
                }
                return dscmd;
            }
        }

        private ExportWorkspaceCommand? ewcmd = null;
        public ExportWorkspaceCommand ExportWorkspaceCmd{
            get
            {
                if (ewcmd == null)
                {
                    ewcmd = new ExportWorkspaceCommand(this);
                }
                return ewcmd;
            }
            }
        #endregion

        #region 函数
        public EditorFrameViewModel()
        {
            documents = new ObservableCollection<PanelViewModel>();
            Documents = new ReadOnlyObservableCollection<PanelViewModel>(documents);
        }

        public void OpenBehaviorTreeView(Datas.BehaviorTree openTree)
        {
            var newDocument = new PanelViewModel(this, openTree);
            this.documents.Add(newDocument);
        }

        public PanelViewModel? FindBehaviorTreeView(Datas.BehaviorTree viewTree)
        {
            for (int i = 0; i < documents.Count; i++)
            {
                if (documents[i].ContentId == viewTree.ID)
                {
                   return documents[i];
                }
            }
            return null;
        }

        public void CloseBehaviorTreeView(Datas.BehaviorTree closeTree)
        {
            for (int i=0;i<documents.Count;i++)
            {
                if (documents[i].ContentId == closeTree.ID)
                {
                    documents.RemoveAt(i);
                    break;
                }
            }
        }

        public void CloseBehaviorTreeViewAll()
        {
            while(documents.Count > 0)
            {
                documents.RemoveAt(0);
            }
        }

        public void OnWorkspaceSelectedItemChangedTree(object sender, RoutedPropertyChangedEventArgs<object> e)
        {
            if (e.NewValue is Datas.BehaviorTree)
            {
                CurrWorkspaceSelectedTree = e.NewValue as Datas.BehaviorTree;
            }
            else
            {
                CurrWorkspaceSelectedTree = null;
            }
        }

        public void OnWorkspaceSelectedNode(object sender, EventArgs e)
        {
            if (sender is string)
            {
                string nodeID = sender as string;
            }
        }
        private void DisplayProperties()
        {
            //TODO: 显示属性
            PropertiesSelectedId = null;
            PropertiesSelectedName = null;
            PropertiesSelectedColor = null;
            PropertiesSelectedDesc = null;
            PropertiesSelectedVisibility = null;
            PropertiesSelectedCategory = null;
            PropertiesSelectedAttributeMap = null;
        }
        #endregion
    }
}
