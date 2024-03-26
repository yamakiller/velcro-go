using Editor.Datas.Files;
using Editor.Panels.Model;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Net.Http.Json;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows;
using System.Xml.Linq;

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
                    /*if (tree.Tree == null)
                    {
                        if (System.IO.File.Exists(filePath))
                        {
                            System.IO.File.Delete(filePath);  
                        }
                        successTotal++;
                        continue;
                    }*/

                    string jsonContent = "";
                    Message = filePath + "[Serializing]";
                    //序列化
                    try
                    {
                        Datas.Files.Behavior3Tree b3tree = new Datas.Files.Behavior3Tree()
                        {
                            ID = tree.ID,
                            Title = tree.Title,
                            Description = tree.Description,
                            Properties = tree.Properties,
                            //Nodes = tree.Nodes,
                        };

                        if (tree.Nodes != null)
                        {
                            b3tree.Nodes = new Dictionary<string, Datas.Files.Behavior3Node>();
                            foreach (var node in tree.Nodes)
                            {
                                Datas.Files.Behavior3Node b3node = new Datas.Files.Behavior3Node()
                                {
                                    ID = node.Value.ID,
                                    Name = node.Value.Name,
                                    Title = node.Value.Title,
                                    Category = node.Value.Category,
                                    Description = node.Value.Description,
                                    Color = node.Value.Color,
                                    //Properties = node.Value.Properties,
                                };
                                if (node.Value.Properties != null)
                                {
                                    b3node.Properties = new Dictionary<string, object>();
                                    foreach (var item in node.Value.Properties)
                                    {
                                        b3node.Properties.Add(item.Value.Key, item.Value.Value);
                                    }
                                }


                                if (node.Value.Children != null)
                                {
                                    b3node.Children = new List<string>();
                                    foreach (var child in node.Value.Children)
                                    {
                                        b3node.Children.Add(child);
                                    }
                                }

                                b3tree.Nodes.Add(node.Key, b3node);
                            }
                        }


                        jsonContent = JsonSerializer.Serialize<Datas.Files.Behavior3Tree>(b3tree);
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
                    catch (Exception ex)
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
                Datas.Files.Workspace currentWorkspaceFilesData = new Datas.Files.Workspace()
                {
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

        public bool? SaveB3(Datas.Workspace workspace, string saveFilePath,string rootTreeID)
        {
            int nTotal = 0, errorTotal = 0, successTotal = 0;
            m_currentWorkspace = workspace;
            List<Datas.Files.B3Tree> currentTrees = new List<Datas.Files.B3Tree>();

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

                    //序列化
                    try
                    {
                        Datas.Files.B3Tree b3tree = new Datas.Files.B3Tree()
                        {
                            ID = tree.ID,
                            Title = tree.Title,
                            Description = tree.Description,
                            Properties = tree.Properties,
                        };

                        if (tree.Nodes != null)
                        {
                            b3tree.Nodes = new Dictionary<string, Datas.Files.B3Node>();
                            foreach (var node in tree.Nodes)
                            {
                                Datas.Files.B3Node b3node = new Datas.Files.B3Node()
                                {
                                    ID = node.Value.ID,
                                    Name = node.Value.Name,
                                    Title = node.Value.Title,
                                    Category = node.Value.Category,
                                    Description = node.Value.Description,
                                    //Properties = node.Value.Properties,
                                };

                                if (node.Key == "root")
                                {
                                    b3tree.Root = node.Value.ID;
                                }

                                if (node.Value.Properties != null)
                                {
                                    b3node.Properties = new Dictionary<string, object>();
                                    foreach (var item in node.Value.Properties)
                                    {
                                        b3node.Properties.Add(item.Value.Key, item.Value.Value);
                                    }
                                }


                                if (node.Value.Children?.Count > 0)
                                {
                                    if (NodeKindConvert.ToKind(node.Value.Category) == NodeKinds.Decorators || NodeKindConvert.ToKind(node.Value.Category) == NodeKinds.Root)
                                    {
                                        b3node.Child = node.Value.Children[0];
                                     }
                                    else
                                    {
                                        b3node.Children = new List<string>();
                                        foreach (var child in node.Value.Children)
                                        {
                                            b3node.Children.Add(child);
                                        }
                                    }
                                }

                                b3tree.Nodes.Add(node.Key, b3node);
                            }
                        }
                        currentTrees.Add(b3tree);
                    }
                    catch (NotSupportedException ex)
                    {
                        errorTotal++;
                        continue;
                    }
                    successTotal++;
                }

                Message = "Save B3File informat";
                //jsonContent = ;
                Datas.Files.B3File currentWorkspaceFilesData = new Datas.Files.B3File()
                {
                    Name = m_currentWorkspace.Name,
                    Data = new Datas.Files.B3Project()
                    {
                        Scope = "",
                        SelectedTree = rootTreeID,
                        Trees = currentTrees.ToArray()
                    },
                };

                try
                {
                    System.IO.File.WriteAllText(saveFilePath,JsonSerializer.Serialize<Datas.Files.B3File>(currentWorkspaceFilesData));

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
                    Message = "Save B3File informat fail " + ex.Message;
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
