using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Editor.Framework;
using Editor.ViewModels;

namespace Editor.Commands
{
    class DebugStopCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public DebugStopCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {

        }
        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            if (contextViewModel.CurrWorkspace == null)
            {
                return;
            }
            contextViewModel.IsWordspaceDebug = false;
            Utils.KafkaConsumer.CallbackEvent -= contextViewModel.OnWorkspaceSelectedNode;
            Utils.KafkaConsumer.StopConsume();
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
