using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Xml.Linq;

namespace Editor.Commands
{
    class NewBehaviorTreeCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public NewBehaviorTreeCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            if (contextViewModel.CurrWorkspace == null)
            {
                return;
            }

            string treeFileName = Utils.RandFileName.GetRandName(contextViewModel.CurrWorkspace.Dir,
                "NewBehaviorTree", ".json");
            if (string.IsNullOrEmpty(treeFileName) )
            {
                Dialogs.WhatDialog.ShowWhatMessage("错误", "创建行为树名失败");
                return;
            }

            var tr = new Datas.BehaviorTree(contextViewModel) 
            { 
                FileName = treeFileName,
                ID = Utils.ShortGuid.Next(),
                Title = treeFileName.Replace(".json", ""),
                Nodes = new Dictionary<string, Datas.BehaviorNode>(),
                Description = ""
            };

            var rootNode = new Datas.BehaviorNode(contextViewModel) {
                ID = Utils.ShortGuid.Next(),
                Name = "root",
                Title = "",
                Description = "",
                Category = "root",
                Color = "#FFB8860B",
            };

            tr.Nodes.Add(rootNode.Name, rootNode);
            contextViewModel.CurrWorkspace.Trees.Add(tr);
            contextViewModel.IsModifyed = true;
            if (!contextViewModel.IsWorkspaceExpanded)
            {
                contextViewModel.IsWorkspaceExpanded = true;
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
