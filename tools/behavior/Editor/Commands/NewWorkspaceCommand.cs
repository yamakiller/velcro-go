using Editor.Datas;
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
            Random rd = new Random();   
            Datas.Workspace workspace = new Datas.Workspace();
            workspace.Name = "BehaviorWorkspace" + rd.Next(1, 100).ToString();
            workspace.ModifyTime = DateTime.Now;
            workspace.CreateTime = DateTime.Now;
            BehaviorTree tree = new BehaviorTree();
         
            tree.id = (new Utils.ShortGuid()).ToString();
            tree.title = "第一个行为树";
            tree.nodes = new Dictionary<string, BehaviorNode>();

            workspace.Trees.Add(tree.id, tree);
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
