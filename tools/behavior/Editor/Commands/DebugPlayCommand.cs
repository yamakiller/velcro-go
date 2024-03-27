using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Editor.Framework;
using Editor.ViewModels;

namespace Editor.Commands
{
    class DebugPlayCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public DebugPlayCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {

        }
        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            if (contextViewModel.CurrWorkspace == null)
            {
                return;
            }
            contextViewModel.IsWordspaceDebug = true;
            for (int i =0; i < contextViewModel.CurrWorkspace.Trees.Count; i++)
            {
                if (contextViewModel.FindBehaviorTreeView(contextViewModel.CurrWorkspace.Trees[i]) == null)
                {
                    contextViewModel.OpenBehaviorTreeView(contextViewModel.CurrWorkspace.Trees[i]);
                }
            }
            Utils.KafkaConsumer.CallbackEvent += contextViewModel.OnWorkspaceSelectedNode;
            Utils.KafkaConsumer.StartConsume("127.0.0.1:9092");
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
