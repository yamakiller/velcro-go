using Editor.Framework;
using Editor.ViewModels;
using Microsoft.Win32;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace Editor.Commands
{
    class OpenWorkspaceCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public OpenWorkspaceCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            OpenFileDialog openFileDlg = new OpenFileDialog();
            openFileDlg.Multiselect = false;
            openFileDlg.Filter = "Behavior Workspace (.json)| *.json";
            if (openFileDlg.ShowDialog() != true)
            {
                return;
            }

            // TODO: 需要关闭久的.
            contextViewModel.Workspace.Clear();
            contextViewModel.Workspace.FilePath = openFileDlg.FileName;
            if (!contextViewModel.Workspace.Load())
            {
                MessageBox.Show("载入工作区文件失败");
                return;
            }

            Dialogs.ScanDialog scanDlg = new Dialogs.ScanDialog();
            scanDlg.Scaning(contextViewModel.Workspace.WorkDir);

            foreach (var file in scanDlg.Files)
            {
                contextViewModel.Workspace.Trees.Add(file);
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
