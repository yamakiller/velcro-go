using Editor.Datas;
using Editor.Framework;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Input;

namespace Editor.ViewModels
{
    class EditorBehaviorViewModel : PaneViewModel
    {
        private EditorFrameViewModel parentViewModel;
        private BehaviorTree btree;
        private PaneCommand closeCommand;

        #region 字段
        public string FileName
        {
            get { return btree.FileName; } 
        }

        public ICommand CloseCommand
        {
            get
            {
                if (closeCommand == null)
                {
                    closeCommand = new PaneCommand((p) => OnClose(), (p) => CanClose());
                }
                return closeCommand;
            }
        }
        #endregion

        #region methods
        private bool CanClose()
        {
            return true;
        }

        private void OnClose()
        {
            parentViewModel.CloseBehaviorTreeView(btree);
        }

        private bool CanSave(object parameter)
        {
            return true;
        }

        private void OnSave(object parameter)
        {
            
        }

        private bool CanSaveAs(object parameter)
        {
            return true;
        }

        private void OnSaveAs(object parameter)
        {

        }
        #endregion

        public EditorBehaviorViewModel(EditorFrameViewModel viewModel, BehaviorTree t)
        {
            parentViewModel = viewModel;

            btree = t;
            ContentId = t.ID;
            Title = t.Title;
        }
    }
}
