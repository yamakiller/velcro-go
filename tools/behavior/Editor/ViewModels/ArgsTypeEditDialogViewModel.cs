using Editor.Framework;
using Editor.Utils;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.ViewModels
{
    class ArgsTypeEditDialogViewModel : ViewModel
    {
        private string? name;

        public string? Name
        {
            get
            {
                return name;
            }
            set
            {
                SetProperty(ref name, value);
            }
        }

        private string? nameError = "";
        public string? NameError
        {
            get
            {
                return nameError;
            }
            set
            {
                SetProperty(ref nameError, value);
            }
        }

        private string? type = "";
        public string? Type
        {
            get
            {
                return type;
            }
            set
            {
                SetProperty(ref type, value);
            }
        }

        private object? defaultValue;
        public object? DefaultValue
        {
            get
            {
                return defaultValue;
            }
            set
            {
                SetProperty(ref defaultValue, value);
            }
        }

        private string? defaultValueError = "";
        public string? DefaultValueError
        {
            get
            {
                return defaultValueError;
            }
            set
            {
                SetProperty(ref defaultValueError, value);
            }
        }

        private string? desc = "";
        public string? Desc
        {
            get
            {
                return desc;
            }
            set
            {
                SetProperty(ref desc, value);
            }
        }
    }
}
