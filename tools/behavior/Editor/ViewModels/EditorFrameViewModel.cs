using Editor.Commands;
using Editor.Framework;

namespace Editor.ViewModels
{
    class EditorFrameViewModel : ViewModel
    {
        #region 属性
        // 是否是只读状态
        bool isReadOnly = false;

        public bool IsReadOnly { get { return isReadOnly; } set { SetProperty(ref isReadOnly, value); } }
        // 是否是修改状态
        bool isModifyed = false;

        public bool IsModifyed { get { return isModifyed; } set { SetProperty(ref isModifyed, value); } }
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
        private Datas.WorkspaceData? wsd = null;
        public Datas.WorkspaceData? CurrWorkspace
        {
            get { return wsd; }
            set 
            {
                if (value == null)
                {
                    IsModifyed = false;
                    Caption = "Behavior Editor";
                } 
                else if (wsd != value)
                {
                    IsModifyed = true;
                    Caption = "*" + value.Dir;
                }

                SetProperty(ref wsd, value); 
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
