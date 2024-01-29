using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Commands
{
    class CloseCurrentWorkspace
    {
        public static void Close(EditorFrameViewModel context)
        {
            if (context.CurrWorkspace != null)
            {
                if (context.IsModifyed)
                {
                    if (Dialogs.WhatDialog.ShowWhatMessage("警告", "当前工作空间未保存,是否需要保存?"))
                    {
                        Dialogs.SaveProccessFrame spf = new Dialogs.SaveProccessFrame();
                        if (spf.Saving(context.CurrWorkspace) == true)
                        {
                            context.IsModifyed = false;
                        }
                    }
                }
                context.CloseBehaviorTreeViewAll();
                context.CurrWorkspace = null;
                context.CurrWorkspaceSelectedTree = null;
                
            }
        }
    }
}
