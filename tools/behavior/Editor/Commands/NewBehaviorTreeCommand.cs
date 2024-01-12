using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows;

namespace Editor.Commands
{
    class NewBehaviorTreeCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public NewBehaviorTreeCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            //contextViewModel.Workspace.WorkDir
            string? filename = Utils.Files.AutoFileNameNumber(contextViewModel.Workspace.WorkDir,
                "NewBehaviorTree", ".json");
            if (string.IsNullOrEmpty(filename) )
            {
                MessageBox.Show("创建行为树文件失败");
                return;
            }

            string filepath = Path.Combine(contextViewModel.Workspace.WorkDir, filename);
            Datas.BehaviorTree behaviorTree = new Datas.BehaviorTree();
            behaviorTree.FilePath = filepath;
            behaviorTree.Sha1 = "";
            behaviorTree.TreeModel = new Datas.Models.BehaviorTreeModel();
            behaviorTree.TreeModel.name = filename.Replace(".json", "");
            behaviorTree.TreeModel.root = new Datas.Models.BehaviorNodeModel() {
                id = Utils.ShortGuid.Next(),
                name = "Sequence",
                desc = "新建行为树",
            };

            contextViewModel.Workspace.Trees.Add(behaviorTree);
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
