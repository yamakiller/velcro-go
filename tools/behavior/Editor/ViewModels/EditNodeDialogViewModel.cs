using Editor.Datas;
using Editor.Framework;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.ViewModels
{
    class EditNodeDialogViewModel : ViewModel
    {
        private string codeFolder;

        public string CodeFolder
        {
            get { return codeFolder; }
            set
            {
                SetProperty(ref codeFolder, value);
            }
        }
       
        public ReadOnlyObservableCollection<Datas.Models.BehaviorNodeTypeModel> Types 
        {
            get;
            set;
        }

        public EditNodeDialogViewModel(ObservableCollection<Datas.Models.BehaviorNodeTypeModel> types)
        {
            Types = new ReadOnlyObservableCollection<Datas.Models.BehaviorNodeTypeModel>(types);
        }
    }
}
