using Editor.Datas;
using System;
using System.Collections;
using System.Collections.Generic;
using System.ComponentModel;
using System.IO;
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
    /// ScanDialog.xaml 的交互逻辑
    /// </summary>
    public partial class ScanDialog : Window, INotifyPropertyChanged
    {
        private List<BehaviorTree> files = new List<BehaviorTree>();

        public List<BehaviorTree> Files { get { return files; } }

        private string _prompt;
        public string Prompt
        {
            get => _prompt;
            set
            {

                _prompt = value;
                RaisePropertyChanged(nameof(Prompt));
            }
        }
        public event PropertyChangedEventHandler PropertyChanged;

        protected void RaisePropertyChanged(string propertyName)
        {
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));
        }

        public ScanDialog()
        {
            InitializeComponent();
            DataContext = this;
        }

        public void Scaning(string workdir)
        {
            //ScanFiles
            var bw = new BackgroundWorker();
            bw.DoWork += delegate
            {
                Prompt = "检索:" + workdir;
                Thread.Sleep(1000);
                Utils.Files.ScanFiles(workdir, new Action<string>(this.fileAccept));
                Prompt = "检索完毕共" + files.Count.ToString() + "个有效文件";
                Thread.Sleep(1000);
            };

            bw.RunWorkerCompleted += delegate
            {
                this.Close();
            };
            bw.RunWorkerAsync();
            ShowDialog();
        }

        private  void fileAccept(string filepath)
        {
            Prompt = filepath;
            if (filepath.EndsWith(".json"))
            {
                try
                {
                    string content = File.ReadAllText(filepath);
                    var model = JsonSerializer.Deserialize<Datas.Models.BehaviorTreeModel>(content);
                    if (model == null) {
                        Prompt = filepath + ":" + "解析错误";
                        return;
                    }

                    files.Add(new BehaviorTree() { FilePath = filepath, TreeModel = model });
                }
                catch(Exception ex) 
                {
                    Prompt = filepath + ":" + ex.Message;
                }
              
            }
        }
    }
}
