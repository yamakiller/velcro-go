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

            string fileName =  Utils.RandFileName.GetRandName(folderDialog.FolderName, "NewWorkspace", ".json");
            if (string.IsNullOrEmpty(fileName)) 
            {
                Dialogs.WhatDialog.ShowWhatMessage("错误", "生成工作空间名失败");
                return;
            }

            CloseCurrentWorkspace.Close(contextViewModel);

            Workspace wkdat = new Workspace(contextViewModel) { 
                Name = fileName.Replace(".json", ""),
                Dir = folderDialog.FolderName, 
                Trees = new System.Collections.ObjectModel.ObservableCollection<BehaviorTree>()};

            contextViewModel.CurrWorkspace = wkdat;
            contextViewModel.IsModifyed = true;
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
