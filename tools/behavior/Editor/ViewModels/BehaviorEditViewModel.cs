using Editor.Commands;
using Editor.Datas;
using Editor.Framework;

namespace Editor.ViewModels
{
    class BehaviorEditViewModel : ViewModel
    {
        bool isReadOnly = false;

        public bool IsReadOnly { get { return isReadOnly; } set { SetProperty(ref isReadOnly, value); } }
    
        Workspace workspace = new Workspace();
    
        public Workspace Workspace { get { return workspace;} 
            set { SetProperty(ref workspace, value); } }

        bool isWorkspaceVaild = false;

        public bool IsWorkspace
        {
            get { return isWorkspaceVaild; }
            set { SetProperty(ref isWorkspaceVaild, value); }
        }

        string caption = "Behavior Editor";

        public string Caption
        {
            get { return caption; }
            set { SetProperty(ref caption, value);}
        }

        private NewWorkspaceCommand _newWorkspaceCmd;

        public NewWorkspaceCommand NewWorkspaceCmd
        {
            get
            {
                if (_newWorkspaceCmd == null)
                {
                    _newWorkspaceCmd = new NewWorkspaceCommand(this);
                }

                return _newWorkspaceCmd;
            }
        }
    }
}
