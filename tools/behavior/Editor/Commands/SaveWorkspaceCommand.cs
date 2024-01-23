using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Commands
{
    class SaveWorkspaceCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public SaveWorkspaceCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
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
