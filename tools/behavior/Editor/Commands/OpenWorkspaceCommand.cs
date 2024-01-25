using Editor.Dialogs;
using Editor.Framework;
using Editor.ViewModels;
using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Xml.Linq;

namespace Editor.Commands
{
    class OpenWorkspaceCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public OpenWorkspaceCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            OpenFileDialog openFileDlg = new OpenFileDialog();
            openFileDlg.Multiselect = false;
            openFileDlg.Filter = "Behavior Workspace (.json)| *.json";
            if (openFileDlg.ShowDialog() != true)
            {
                return;
            }

            string jsonContent = System.IO.File.ReadAllText(openFileDlg.FileName);
            if (string.IsNullOrEmpty(jsonContent))
            {
                Dialogs.WhatDialog.ShowWhatMessage("错误", "open workspace file fail does not exist");
                return;
            }


            Datas.Files.Workspace? wrk = null;

            try
            {
                wrk = JsonSerializer.Deserialize<Datas.Files.Workspace>(jsonContent);
            }
            catch (Exception ex) 
            {
                Dialogs.WhatDialog.ShowWhatMessage("错误", "open workspace file fail " + ex.Message);
                return;
            }

            if (wrk == null)
            {
                return;
            }

            

            // 载入行为树数据
            OpenProccessFrame openProcFame = new OpenProccessFrame();
            if (openProcFame.Openning(wrk) != true)
            {
                return;
            }

            Datas.Workspace currentWorkspace = new Datas.Workspace()
            {
                Name = wrk.Name,
                Dir = wrk.Dir,
                Trees = new System.Collections.ObjectModel.ObservableCollection<Datas.BehaviorTree>()
            };

            foreach(Datas.Files.Behavior3Tree tree in openProcFame.Trees)
            {
                currentWorkspace.Trees.Add(new Datas.BehaviorTree() { FileName = tree.Title + ".json", Tree = tree });
            }

            CloseCurrentWorkspace.Close(contextViewModel);
            contextViewModel.CurrWorkspace = currentWorkspace;
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
