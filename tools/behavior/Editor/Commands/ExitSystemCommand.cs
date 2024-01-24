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
            if (contextViewModel.IsModifyed)
            {
               if (Dialogs.WhatDialog.ShowWhatMessage("警告", "当前工作空间未保存,是否需要保存?"))
                {
                    // TODO: 保存
                }
            }

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
