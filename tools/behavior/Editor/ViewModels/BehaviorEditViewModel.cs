using Editor.Commands;
using Editor.Datas;
using Editor.Framework;
using System.Collections.ObjectModel;
using System.IO;

namespace Editor.ViewModels
{
    class BehaviorEditViewModel : ViewModel
    {
        bool isReadOnly = false;

        public bool IsReadOnly { get { return isReadOnly; } set { SetProperty(ref isReadOnly, value); } }
    
        Workspace workspace = new Workspace();
    
        public Workspace Workspace 
        {
            get { return workspace;} 
            set { SetProperty(ref workspace, value); } 
        }

        bool isWorkspaceVaild = false;

        public bool IsWorkspace
        {
            get { return isWorkspaceVaild; }
            set 
            {   if (value)
                {
                    Caption = "Workspace:" + workspace.WorkDir;
                }
                else
                {
                    Caption = "BehaviorEditor";
                }
                SetProperty(ref isWorkspaceVaild, value); 
            }
        }

        bool isWorkspaceModify = false;
        public bool IsWorkspaceModify
        {
            get { return isWorkspaceModify; }
            set 
            {   if (value && IsWorkspace)
                {
                    Caption = "*Workspace:" + Workspace.WorkDir;
                }
                else if (!value && IsWorkspace) 
                {
                    Caption = "Workspace:" + Workspace.WorkDir;
                }
                
                SetProperty(ref isWorkspaceModify, value); 
            }
        }

        string caption = "Behavior Editor";

        public string Caption
        {
            get { return caption; }
            set { SetProperty(ref caption, value);}
        }
       
        public ReadOnlyObservableCollection<BehaviorTree> Trees {
            get;
            set;
        }

        private NewWorkspaceCommand? newWorkspaceCmd = null;

        public NewWorkspaceCommand NewWorkspaceCmd
        {
            get
            {
                if (newWorkspaceCmd == null)
                {
                    newWorkspaceCmd = new NewWorkspaceCommand(this);
                }

                return newWorkspaceCmd;
            }
        }

        private ExitSystemCommand? exitSystemCmd = null;

        public ExitSystemCommand ExitSystemCmd
        {
            get
            {
                if (exitSystemCmd == null)
                {
                    exitSystemCmd = new ExitSystemCommand(this);
                }

                return exitSystemCmd;
            }
        }

        private NewBehaviorTreeCommand? newBehaviorTreeCmd = null;
        public NewBehaviorTreeCommand NewBehaviorTreeCmd
        {
            get
            {
                if (newBehaviorTreeCmd == null)
                {
                    newBehaviorTreeCmd = new NewBehaviorTreeCommand(this);
                }

                return newBehaviorTreeCmd;
            }
        }

        private SaveWorkspaceCommand? saveWorkspaceCmd = null;
        public SaveWorkspaceCommand SaveWorkspaceCmd
        {
            get
            {
                if (saveWorkspaceCmd == null)
                {
                    saveWorkspaceCmd = new SaveWorkspaceCommand(this);
                }
                return saveWorkspaceCmd;
            }
        }

        private OpenNodeEditorViewCommand? openNodeEditorViewCmd = null;
        public OpenNodeEditorViewCommand OpenNodeEditorViewCmd
        {
            get
            {
                if (openNodeEditorViewCmd == null)
                {
                    openNodeEditorViewCmd= new OpenNodeEditorViewCommand(this);
                }
                return openNodeEditorViewCmd;
            }
        }

        private OpenEditDefaultNodeCommand? openDefaultEditorViewCmd = null;
        public OpenEditDefaultNodeCommand OpenDefaultEditorViewCmd
        {
            get
            {
                if (openDefaultEditorViewCmd == null)
                {
                    openDefaultEditorViewCmd = new OpenEditDefaultNodeCommand(this);
                }

                return openDefaultEditorViewCmd;
            }
        }

        public BehaviorEditViewModel()
        {
            Trees = new ReadOnlyObservableCollection<BehaviorTree>(workspace.Trees);

        }
    }
}
