using Editor.Framework;
using Editor.ViewModels;

namespace Editor.Commands
{
    class NewWorkspaceCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public NewWorkspaceCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
        }

        public override bool CanExecute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            return true;
        }
    }
}
