﻿using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Commands
{
    class DelBehaviorTreeCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public DelBehaviorTreeCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            int idx = (int)parameter;
            if (idx < 0)
            {
                return;
            }


            if (idx >= contextViewModel.Workspace.Trees.Count) {
                return;
            }

            contextViewModel.Workspace.Trees.RemoveAt(idx);
            // 关闭视图
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
