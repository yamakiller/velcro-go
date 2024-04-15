using Editor.Dialogs;
using Editor.Framework;
using Editor.ViewModels;
using MaterialDesignThemes.Wpf;
using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Editor.Commands
{

    class ExportWorkspaceCommand : ViewModelCommand<EditorFrameViewModel>
    {
        public ExportWorkspaceCommand(EditorFrameViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(EditorFrameViewModel contextViewModel, object parameter)
        {
            if (contextViewModel.CurrWorkspace == null)
            {
                return;
            }

            SaveFileDialog saveFileDialog = new SaveFileDialog();
            saveFileDialog.Filter = "行为树文件(*.b3, *.B3)|*.b3;*.B3|所有文件(*.*)|*.*";
            saveFileDialog.Title = "请选择文件保存文件夹";

            if (saveFileDialog.ShowDialog() != true)
            {
                return;
            }

            SelectedRootTreeDialog selectedRoot = new SelectedRootTreeDialog();
            if (selectedRoot.ShowTrees(contextViewModel.CurrWorkspace.Trees) != true)
            {
                return;
            }
            Dialogs.SaveProccessFrame spf = new Dialogs.SaveProccessFrame();
            if (spf.SaveB3(contextViewModel.CurrWorkspace, saveFileDialog.FileName, selectedRoot.SelectedRoot.Name) == true)
            {
                contextViewModel.IsModifyed = false;
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
