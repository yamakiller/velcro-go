using Editor.Dialogs;
using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Commands
{
    class OpenNodeEditorViewCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public OpenNodeEditorViewCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            EditNodeDialog editNodeDialog = new EditNodeDialog(contextViewModel.Workspace.Types.ToArray());
            editNodeDialog.ShowDialog();
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
