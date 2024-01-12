using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Editor.Commands
{
    class SaveWorkspaceCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public SaveWorkspaceCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            if (contextViewModel.Workspace.Save(null))
            {
                contextViewModel.IsWorkspaceModify = false;
            }
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
