using Editor.Framework;
using Editor.ViewModels;
using System.Windows;

namespace Editor.Commands
{
    class ExitSystemCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public ExitSystemCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            Application.Current.Shutdown();
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
