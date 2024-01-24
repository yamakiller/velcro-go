using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Shapes;

namespace Editor.Dialogs
{
    /// <summary>
    /// OpenProccessFrame.xaml 的交互逻辑
    /// </summary>
    public partial class OpenProccessFrame : Window, INotifyPropertyChanged
    {
        private List<Datas.Files.Behavior3Tree> trees = new List<Datas.Files.Behavior3Tree>();
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

        public List<Datas.Files.Behavior3Tree> Trees { get => trees; }



        public event PropertyChangedEventHandler? PropertyChanged;

        protected void RaisePropertyChanged(string propertyName)
        {
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
        }

        public OpenProccessFrame()
        {
            WindowStartupLocation = WindowStartupLocation.CenterScreen;
            InitializeComponent();
            DataContext = this;
        }


        public bool? Openning(Datas.Files.Workspace workspace)
        {
            bool reuslt = false;
            var bw = new BackgroundWorker();
            bw.DoWork += delegate
            {
                Message = "Work directory " + workspace.Dir;
                Thread.Sleep(1000);
                if (workspace.Files == null)
                {
                    return;
                }

                foreach (var file in workspace.Files)
                {
                    string jsonContext = "";
                    string filePath = System.IO.Path.Combine(workspace.Dir, file);
                    Message = "load " + filePath;
                    try
                    {
                        jsonContext = System.IO.File.ReadAllText(filePath);
                    }
                    catch(Exception ex)
                    {
                        Message += ex.ToString();
                        Thread.Sleep(2000);
                        continue;
                    }
                    Datas.Files.Behavior3Tree? tree = null;
                    try
                    {
                        Message = filePath + "[Deserialize]";
                        tree = JsonSerializer.Deserialize<Datas.Files.Behavior3Tree>(jsonContext);
                    }
                    catch (Exception ex)
                    {
                        Message += ex.ToString();
                        Thread.Sleep(2000);
                        continue;
                    }

                    if (tree == null)
                    {
                        continue;
                    }

                    trees.Add(tree);
                }
                reuslt = true;
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
