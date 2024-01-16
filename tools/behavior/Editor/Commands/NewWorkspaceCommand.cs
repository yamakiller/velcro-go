using Editor.Datas;
using Editor.Framework;
using Editor.ViewModels;
using Microsoft.Win32;
using System.IO;
using System.Windows;

namespace Editor.Commands
{
    class NewWorkspaceCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public NewWorkspaceCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            var newWorkspace = new Dialogs.CreateWorkspaceDialog(); 
            var result = newWorkspace.ShowDialog();
            if (result != true)
            {
                return;
            }

            var filepath = Path.Combine(newWorkspace.WorkspaceFolder,
                newWorkspace.WorkspaceName + ".json");

            if (File.Exists(filepath))
            {
               if (MessageBoxResult.Cancel == 
                   MessageBox.Show("文件:" + filepath + "已存在?", "警告", MessageBoxButton.OKCancel))
               {
                    return;
               }
            }

            contextViewModel.Workspace.FilePath = filepath;
            contextViewModel.Workspace.WorkDir = newWorkspace.WorkspaceFolder;

            // 生成默认节点配置文件

            Dialogs.ScanDialog scanDlg = new Dialogs.ScanDialog();
            scanDlg.Scaning(contextViewModel.Workspace.WorkDir);

            foreach(var file in scanDlg.Files)
            {
                contextViewModel.Workspace.Trees.Add(file);
            }
            
            contextViewModel.IsWorkspace = true;
            contextViewModel.IsWorkspaceModify = true;
        }

        public override bool CanExecute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            if (contextViewModel.IsReadOnly)
            {
                return false;
            }
            return true;
        }
    }
}
