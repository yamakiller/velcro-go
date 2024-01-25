using Editor.Commands;
using Editor.Framework;

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
        #endregion

        /// <summary>
        /// 窗口标题
        /// </summary>
        private string caption = "Behavior Editor";
        public string Caption
        {
            get { return caption; }
            set { SetProperty(ref caption, value); }
        }

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

        #endregion
    }
}
