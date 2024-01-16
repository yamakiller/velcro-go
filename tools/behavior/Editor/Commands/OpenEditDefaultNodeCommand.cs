using Editor.Configs;
using Editor.Dialogs;
using Editor.Framework;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows;

namespace Editor.Commands
{
    class OpenEditDefaultNodeCommand : ViewModelCommand<BehaviorEditViewModel>
    {
        public OpenEditDefaultNodeCommand(BehaviorEditViewModel contextViewModel) : base(contextViewModel)
        {
        }

        public override void Execute(BehaviorEditViewModel contextViewModel, object parameter)
        {
            // 打开默认节点配置文件
            List<Datas.Models.BehaviorNodeTypeModel> array = new List<Datas.Models.BehaviorNodeTypeModel>();
            string defaultNodeFilePath = Path.Combine(System.AppDomain.CurrentDomain.BaseDirectory, Configs.Config.DefaultNodeFile);

            string defaultJson = "";

            try
            {
                defaultJson = File.ReadAllText(defaultNodeFilePath);
            }
            catch(Exception)
            {

            }
            
            if (!string.IsNullOrEmpty(defaultJson))
            {
                Datas.Models.BehaviorNodeTypeModel[]? tmps =  JsonSerializer.Deserialize<Datas.Models.BehaviorNodeTypeModel[]>(defaultJson);
                if (tmps != null)
                {
                    array.AddRange(tmps);
                }
            }

            EditNodeDialog editNodeDialog = new EditNodeDialog(array.ToArray());
            var result =  editNodeDialog.ShowDialog();
            if (result != true)
            {
                return;
            }

            array.Clear();
            Datas.Models.BehaviorNodeTypeModel[] newArray = editNodeDialog.GetTypes();
            var options = new JsonSerializerOptions { WriteIndented = true };
            defaultJson = JsonSerializer.Serialize<Datas.Models.BehaviorNodeTypeModel[]>(newArray, options);

            if (!Path.Exists(Path.GetDirectoryName(defaultNodeFilePath)))
            {
                Directory.CreateDirectory(Path.GetDirectoryName(defaultNodeFilePath));
            }

            File.WriteAllText(defaultNodeFilePath, defaultJson);
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
