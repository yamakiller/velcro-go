using Editor.Datas.Files;
using Editor.Framework;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Datas
{
    public class BehaviorNode
    {
        private ViewModelData contextViewModel;

        private string id = "";
        public required string ID { get { return id; } 
            
            set
            {
                string oldId = id;
                id = value;
                if (contextViewModel != null && oldId != value)
                {
                    contextViewModel.IsModifyed = true;
                }
            }
        }
        private string name = "";
        public required string Name { get { return name; } 
            set
            {
                string oldName = name;
                name = value;
                if (contextViewModel != null && oldName != value)
                {
                    contextViewModel.IsModifyed = true;
                }
            }
        }
        private string category = "";
        public required string Category { get { return category; }
            set
            {
                string oldCategory = category;
                category = value;
                if (contextViewModel != null && oldCategory != value)
                {
                    contextViewModel.IsModifyed = true;
                }
            }
        }
        private string title = "";
        public required string Title { get { return title; }
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
        public required string Description { get { return description; } 
            
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
        public ObservableCollection<string>? Children { get; set; }

        private Dictionary<string, object>? properties;
        public Dictionary<string, object>? Properties
        {
            get { return properties; }
            set
            {
                Dictionary<string, object>? old = properties;
                properties = value;
                if (contextViewModel != null && old != value)
                { 
                    contextViewModel.IsModifyed = true;
                }
            }
        }

        private string color = "";
        public string Color
        {
            get { return color; }
            set
            {
                string oldColor = color;
                color = value;
                if (contextViewModel != null && oldColor != value)
                {
                    contextViewModel.IsModifyed = true;
                }
            }
        }

        public BehaviorNode(ViewModelData model)
        {
            contextViewModel = model;
        }
    }
}
