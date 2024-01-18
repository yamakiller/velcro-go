using Editor.Datas;
using Editor.Datas.Models;
using Editor.Framework;
using Editor.Utils;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Controls.Primitives;

namespace Editor.ViewModels
{
    class EditNodeDialogViewModel : ViewModel
    {

        private BehaviorNodeTypeModel selectedNode;

        public BehaviorNodeTypeModel SelectedNode
        {
            get { return selectedNode; }
            set
            {
                if (selectedNode != value)
                {
                    SetProperty(ref selectedNode, value);
                    RaisePropertyChanged("SelectedNodeFile");
                    RaisePropertyChanged("SelectedNodeName");
                    RaisePropertyChanged("SelectedNodePackageName");
                    RaisePropertyChanged("SelectedNodeType");
                    RaisePropertyChanged("SelectedNodeDesc");
                    RaisePropertyChanged("SelectedNodeInput");
                    RaisePropertyChanged("SelectedNodeOutput");
                    RaisePropertyChanged("SelectedNodeArgs");
                    RaisePropertyChanged("SelectedNodeCode");
                }
            }
        }

        public string? SelectedNodeFile
        {
            get
            {
                if (selectedNode == null)
                {
                    return "";
                }

                return selectedNode.src;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNode.src = value;
                    RaisePropertyChanged("SelectedNodeFile");
                    if (value != null)
                    {
                        NodeAnalysis analysis = new NodeAnalysis(value);
                        if (analysis.Do())
                        {
                            SelectedNodeName = analysis.Name;
                            SelectedNodePackageName = analysis.Package;
                            SelectedNodeCode = analysis.Content;
                        }
                    }
                }
            }
        }

        public string SelectedNodeName
        {
            get 
            {
                if (selectedNode == null)
                {
                    return "";
                }

                return selectedNode.name;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNode.name = value;
                    foreach(BehaviorNodeTypeModel t in types)
                    {
                        if (t == selectedNode)
                        {
                            t.name = value;
                        }
                    }
                    RaisePropertyChanged("SelectedNodeName");
                    RaisePropertyChanged("SelectedNode");
                }
            }
        }

        private string selectedNodePackageName;
        public string SelectedNodePackageName
        {
            get
            {
                if (selectedNode == null)
                {
                    return "";
                }

                return selectedNodePackageName;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNodePackageName = value;
                    RaisePropertyChanged("SelectedNodePackageName");
                }
            }
        }

        public string? SelectedNodeType
        {
            get
            {
                if (selectedNode == null)
                {
                    return "";
                }

                return selectedNode.type;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNode.type = value;
                    RaisePropertyChanged("SelectedNodeType");
                }
            }
        }

        public string? SelectedNodeDesc
        {
            get
            {
                if (selectedNode == null)
                {
                    return "";
                }

                return selectedNode.desc;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNode.desc = value;
                    RaisePropertyChanged("SelectedNodeDesc");
                }
            }
        }


        public string? SelectedNodeInput
        {
            get
            {
                if (selectedNode == null)
                {
                    return "";
                }

                return selectedNode.input;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNode.input = value;
                    RaisePropertyChanged("SelectedNodeInput");
                }
            }
        }

        public string? SelectedNodeOutput
        {
            get
            {
                if (selectedNode == null)
                {
                    return "";
                }

                return selectedNode.input;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNode.input = value;
                    RaisePropertyChanged("SelectedNodeOutput");
                }
            }
        }

        public ObservableCollection<ArgsDefType>? SelectedNodeArgs
        {
            get
            {
                if (selectedNode == null)
                {
                    return null;
                }

                return selectedNode.args;
            }
            set
            {
                if (selectedNode != null)
                {
                    selectedNode.args = value;
                    RaisePropertyChanged("SelectedNodeArgs");
                }
            }
        }

        private string selectedNodeCode;
        public string SelectedNodeCode
        {
            get
            {
                if (selectedNode == null)
                {
                    return string.Empty;
                }
                    
                return selectedNodeCode;
            }

            set
            {
                if (selectedNode == null)
                {
                    return;
                }

                selectedNodeCode = value;
                RaisePropertyChanged("SelectedNodeCode");
            }
        }

        private ObservableCollection<BehaviorNodeTypeModel> types;

        public ObservableCollection<BehaviorNodeTypeModel> Types 
        {
            get { return types; }
            set { types = value; }
        }

        public EditNodeDialogViewModel(ObservableCollection<BehaviorNodeTypeModel> s)
        {
            types = s;
        }
    }
}
