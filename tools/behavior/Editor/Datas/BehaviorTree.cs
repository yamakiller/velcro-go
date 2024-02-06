using Editor.Datas.Files;
using Editor.Framework;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class BehaviorTree
    {
        private ViewModelData contextViewModel;
        private string fileName = "";
        public required string FileName 
        { 
            get { return fileName; } 
            
            set {
                string oldFileName= fileName;
                fileName = value;
                if (contextViewModel != null && oldFileName != value)
                {
                    contextViewModel.IsModifyed = true;
                }
            } 
        }

        public required string ID { get; set; }
        private string title = "";
        public required string Title 
        {
            get { return title; } 
            set
            {
                string oldTitle = title;
                title = value;
                if (contextViewModel != null && oldTitle != value)
                {
                    contextViewModel.IsModifyed = true;
                }
            }
        }

        private string description = "";
        public required string Description
        {
            get { return description; }
            set
            {
                string oldDesc = description;
                description = value;
                if (contextViewModel != null && oldDesc != value)
                {
                    contextViewModel.IsModifyed = true;
                }
            }
        }

        public Dictionary<string, object>? Properties { get; set; }

        public Dictionary<string, BehaviorNode> Nodes { get; set; }

        public BehaviorTree(ViewModelData model)
        {
            contextViewModel = model;
        }
    }
}
