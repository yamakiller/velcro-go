using Editor.Datas;
using Editor.Framework;
using Editor.ViewModels;
using Microsoft.Win32;

namespace Editor.Commands
{
    class NewWorkspaceCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public NewWorkspaceCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            var folderDialog = new OpenFolderDialog() { 
                Title = "Workspace",
                InitialDirectory = Environment.GetFolderPath(Environment.SpecialFolder.Personal),
                Multiselect = false, };
            var result = folderDialog.ShowDialog();
            if (result != true)
            {
                return;
            }

            contextViewModel.Workspace.WorkDir = folderDialog.FolderName;

            Dialogs.ScanDialog scanDlg = new Dialogs.ScanDialog();
            scanDlg.Scaning(contextViewModel.Workspace.WorkDir);

            foreach(var file in scanDlg.Files)
            {
                contextViewModel.Workspace.AddBT(file);
            }
            contextViewModel.Caption = "Workspace:" + contextViewModel.Workspace.WorkDir;
            contextViewModel.IsWorkspace = true;
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
