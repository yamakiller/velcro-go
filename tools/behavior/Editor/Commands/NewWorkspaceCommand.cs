using Editor.Framework;
using Editor.ViewModels;
using Editor.Datas;
using Microsoft.Win32;

namespace Editor.Commands
{
    class NewWorkspaceCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public NewWorkspaceCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            var folderDialog = new OpenFolderDialog()
            {
                Title = "Workspace",
                InitialDirectory = Environment.GetFolderPath(Environment.SpecialFolder.Personal),
                Multiselect = false,
            };

            var result = folderDialog.ShowDialog();
            if (result != true)
            {
                return;
            }

            WorkspaceData wks = new WorkspaceData() { Dir = folderDialog.FolderName, 
                Files = null };

            // TODO: 如果存在久的工作空间，哪么将其关闭

            contextViewModel.CurrWorkspace = wks;
        }

        public override bool CanExecute(EditorFrameViewModel contextViewModel, object parameter)
        {
            if (contextViewModel.IsReadOnly)
            {
                return false;
            }
            return true;
        }
    }
}
