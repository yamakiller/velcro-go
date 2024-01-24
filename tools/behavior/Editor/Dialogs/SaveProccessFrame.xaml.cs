﻿using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Net.Http.Json;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows;

namespace Editor.Dialogs
{
    /// <summary>
    /// SaveProccessFrame.xaml 的交互逻辑
    /// </summary>
    public partial class SaveProccessFrame : Window, INotifyPropertyChanged
    {
        /// <summary>
        /// 当前要保存的空间数据
        /// </summary>
        private Datas.Workspace m_currentWorkspace;

        /// <summary>
        /// 提示信息
        /// </summary>
        private string m_message;

        public string Message
        {
            get => m_message;
            set
            {

                m_message = value;
                RaisePropertyChanged(nameof(Message));
            }
        }


        public event PropertyChangedEventHandler? PropertyChanged;

        protected void RaisePropertyChanged(string propertyName)
        {
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
        }

        public SaveProccessFrame()
        {
            WindowStartupLocation = WindowStartupLocation.CenterScreen;
            InitializeComponent();
            DataContext = this;
        }

        public bool? Saving(Datas.Workspace workspace)
        {
            int nTotal = 0, errorTotal = 0, successTotal = 0;
            m_currentWorkspace = workspace;
            List<string> currentTrees = new List<string>();

            bool reuslt = false;
            var bw = new BackgroundWorker();
            bw.DoWork += delegate
            {
                Message = "Work directory " + m_currentWorkspace.Dir;
                Thread.Sleep(1000);
                nTotal = m_currentWorkspace.Trees.Count;
                foreach (var tree in m_currentWorkspace.Trees)
                {
                    if (string.IsNullOrEmpty(tree.FileName))
                    {
                        successTotal++;
                        continue;
                    }
                    string filePath = System.IO.Path.Combine(m_currentWorkspace.Dir, tree.FileName);
                    Message = filePath;
                    if (tree.Tree == null)
                    {
                        if (System.IO.File.Exists(filePath))
                        {
                            System.IO.File.Delete(filePath);  
                        }
                        successTotal++;
                        continue;
                    }

                    string jsonContent = "";
                    Message = filePath + "[Serializing]";
                    //序列化
                    try
                    {
                        jsonContent = JsonSerializer.Serialize<Datas.Files.Behavior3Tree>(tree.Tree);
                    }
                    catch (NotSupportedException ex) 
                    {
                        Message = filePath + "[Serialize fail " + ex.Message + "]";
                        errorTotal++;
                        continue;
                    }
                    

                    Message = filePath + "[Serialize Complate]";

                    try
                    {
                        System.IO.File.WriteAllText(filePath, jsonContent);
                    }
                    catch(Exception ex) 
                    {
                        Message = filePath + ex.Message;
                        errorTotal++;
                        continue;
                    }

                    currentTrees.Add(tree.FileName);
                    successTotal++;
                }

                Message = "Save Workspace informat";
                //jsonContent = ;
                Datas.Files.Workspace currentWorkspaceFilesData = new  Datas.Files.Workspace(){
                    Name = m_currentWorkspace.Name,
                    Dir = m_currentWorkspace.Dir,
                    Files = currentTrees.ToArray(),
                };

                try
                {
                    System.IO.File.WriteAllText(System.IO.Path.Combine(m_currentWorkspace.Dir, m_currentWorkspace.Name + ".json"),
                        JsonSerializer.Serialize<Datas.Files.Workspace>(currentWorkspaceFilesData));

                    Message = "Save Statistics Total:" + nTotal.ToString() +
                        ", Error:" + errorTotal.ToString() +
                        ", Success:" + successTotal.ToString();
                    Thread.Sleep(1000);
                    Message = "Save Success";
                    Thread.Sleep(2000);
                    reuslt = true;
                }
                catch (Exception ex) 
                {
                    Message = "Save Workspace informat fail " + ex.Message;
                    Thread.Sleep(2000);
                }
            };


            bw.RunWorkerCompleted += delegate
            {
                this.DialogResult = reuslt;
                this.Close();
            };
            bw.RunWorkerAsync();
            return ShowDialog();
        }
    }
}
