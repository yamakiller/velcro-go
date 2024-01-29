using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Commands
{
    class OpenTreeCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public OpenTreeCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            if (parameter == null)
            {
                return;
            }

            if (contextViewModel.FindBehaviorTreeView(parameter as Datas.BehaviorTree) == null)
            {
                contextViewModel.OpenBehaviorTreeView(parameter as Datas.BehaviorTree);
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
