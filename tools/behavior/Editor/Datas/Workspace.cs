using Editor.Framework;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class Workspace
    {
        private ViewModelData contextViewModel;

        private string name = "";
        public required string Name 
        { 
            get { return name; }
            set
            {
                string oldName = name;
                name = value;
                if (contextViewModel != null && oldName != name)
                {
                    contextViewModel.IsModifyed = true;
                }
            }
        }
        public required string Dir { get; set; }
        public required ObservableCollection<BehaviorTree> Trees { get; set; }

        public Workspace(ViewModelData model)
        {
            contextViewModel = model;
        }
    }
}
