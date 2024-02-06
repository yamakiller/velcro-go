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

            Datas.Workspace currentWorkspace = new Datas.Workspace(contextViewModel)
            {
                Name = wrk.Name,
                Dir = wrk.Dir,
                Trees = new System.Collections.ObjectModel.ObservableCollection<Datas.BehaviorTree>()
            };

            foreach(Datas.Files.Behavior3Tree b3tree in openProcFame.Trees)
            {
                Datas.BehaviorTree tree  = new Datas.BehaviorTree(contextViewModel) {
                    FileName = b3tree.Title + ".json",
                    ID = b3tree.ID,
                    Title = b3tree.Title,
                    Description = b3tree.Description,
                    Properties = b3tree.Properties,
                };

                if (b3tree.Nodes != null)
                {
                    tree.Nodes = new Dictionary<string, Datas.BehaviorNode>();
                    foreach ( var b3node in b3tree.Nodes)
                    {
                        Datas.BehaviorNode node = new Datas.BehaviorNode(contextViewModel)
                        {
                            ID = b3node.Value.ID,
                            Name = b3node.Value.Name,
                            Category = b3node.Value.Category,
                            Title = b3node.Value.Title,
                            Description = b3node.Value.Description,
                            Color = b3node.Value.Color,
                            Properties = b3node.Value.Properties,
                        };

                        if (b3node.Value.Children != null)
                        {
                            node.Children = new System.Collections.ObjectModel.ObservableCollection<string>(b3node.Value.Children);
                        }
                        
                        tree.Nodes.Add(b3node.Key, node);
                        
                    }
                }

               
                currentWorkspace.Trees.Add(tree);
            }

            CloseCurrentWorkspace.Close(contextViewModel);
            contextViewModel.CurrWorkspace = currentWorkspace;
            
            contextViewModel.IsModifyed = false;
            if (contextViewModel.CurrWorkspace.Trees != null && 
                contextViewModel.CurrWorkspace.Trees.Count > 0)
            {
                contextViewModel.OpenBehaviorTreeView(contextViewModel.CurrWorkspace.Trees[0]);
                contextViewModel.CurrWorkspaceSelectedTree = contextViewModel.CurrWorkspace.Trees[0];
            }
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
