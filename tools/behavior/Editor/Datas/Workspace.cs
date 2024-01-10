using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Reflection.Metadata;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows.Shapes;

namespace Editor.Datas
{

    //public class Workspace
    public  class Workspace
    {
        private string m_filepath = "";     // 文件路径
        private string m_nodeConfPath = ""; // 节点配置文件路径
        private string m_workdir = "";      // 工作目录
        private Models.WorkspaceModel? m_model;
        private Dictionary<string, Models.BehaviorNodeTypeModel> m_name2conf = new Dictionary<string, Models.BehaviorNodeTypeModel>();
        private Models.BehaviorNodeTypeModel[]? m_type;
        private List<BehaviorTree> m_treeFilePath = new List<BehaviorTree>();

        private string m_lastError;

        public string NodeConfPath
        {
            get { return m_nodeConfPath; }
            set { m_nodeConfPath = value; }
        }

        public string WorkDir
        {
            get { return m_workdir; }
            set { m_workdir = value; }
        }

        public Models.WorkspaceModel GetModel()
        {
            return m_model;
        }

        public string GetLastError()
        {
            return m_lastError;
        }

        public bool Load()
        {
            if (m_filepath == null)
            {
                return false;
            }

            string content = File.ReadAllText(m_filepath);
            if (content == null)
            {
                return false;
            }

            try
            {
               var model = JsonSerializer.Deserialize<Models.WorkspaceModel>(content);
                if (model == null)
                {
                    return false;
                }
                if (model.IsRelative != null && model.IsRelative == true)
                {
                    var root = System.IO.Path.GetDirectoryName(m_filepath);
                    Debug.Assert(root != null);
                    m_nodeConfPath = System.IO.Path.Combine(root, model.NodeConfPath);
                    m_workdir = System.IO.Path.Combine(root, model.WorkDir);
                } 
                else
                {
                    m_nodeConfPath = model.NodeConfPath;
                    m_workdir = model.WorkDir;
                }
                m_model = model;
            }
            catch (JsonException ex) {
                m_lastError = ex.Message;
                return false;
            }
            
            return true;
        }

        public bool Save(string? filepath) 
        { 
            if (filepath != null)
            {
                m_filepath = filepath;
            }
            if (string.IsNullOrEmpty(m_filepath))
            {
                return false;
            }


            var options = new JsonSerializerOptions { WriteIndented = true };
            File.WriteAllText(m_filepath, 
                JsonSerializer.Serialize<Models.WorkspaceModel>(new Models.WorkspaceModel() { NodeConfPath = m_nodeConfPath, WorkDir = m_workdir }, options));

            return true;
        }

        /// <summary>
        /// 添加一棵树
        /// </summary>
        /// <param name="behaviorTree"></param>
        public void AddBT(BehaviorTree behaviorTree)
        {
            m_treeFilePath.Add(behaviorTree);
        }

        /// <summary>
        /// 删除一颗树
        /// </summary>
        /// <param name="index"></param>
        public void RemoveBT(int index ) {
            m_treeFilePath.RemoveAt(index);
        }

        private bool initNodeConf()
        {
            if (string.IsNullOrEmpty(m_nodeConfPath))
            {
                return false;
            }

            var types = JsonSerializer.Deserialize<Models.BehaviorNodeTypeModel[]>(File.ReadAllText(m_nodeConfPath));
            Debug.Assert(types != null);

            m_name2conf = new Dictionary<string, Models.BehaviorNodeTypeModel>();
            foreach(var t in types)
            {
                m_name2conf.Add(t.name, t);
            }
            m_type = types;
            return true;
        }
    }
}
