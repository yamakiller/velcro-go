using Editor.Framework;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Dialogs
{
    class WhatDialogViewModel : ViewModel
    {
        private string caption = "";
        public string Caption 
        { 
            get
            {
                return caption;
            }

            set
            {
                SetProperty(ref caption, value);
            }
        }

        private string message = "";
        public string Message
        {
            get
            {
                return message;
            }

            set
            {
                SetProperty(ref message, value);
            }
        }
    }
}
