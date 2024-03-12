using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Panels.Model
{
    class BNode : INotifyPropertyChanged
    {

        #region 属性
        [Browsable(false)]
        public NodeKinds Kind { get; private set; }

        private int m_column;
        public int Column { get { return m_column; } set { m_column = value; OnPropertyChanged("Column"); } }
        private float m_row;
        public float Row { get { return m_row; } set { m_row = value; OnPropertyChanged("Row"); } }

        private int m_width;
        public int Width { get { return m_width; } set { m_width = value; OnPropertyChanged("Width"); } }

        private int m_height;
        public int Hegith { get { return m_height; } set { m_height = value; OnPropertyChanged("Height"); } }

        private string m_id = "";
        public string Id
        {
            get { return m_id; }
            set
            {
                m_id = value;
                OnPropertyChanged("Id");
            }
        }

        private string m_name = "";
        public string Name
        {
            get { return m_name; }
            set
            {
                m_name = value;
                OnPropertyChanged("Name");
            }
        }

        private string m_category = "";
        public string Category
        {
            get { return m_category; }
            set
            {
                m_category = value;
                Kind = NodeKindConvert.ToKind(value);
                OnPropertyChanged("Category");
            }
        }

        private string m_title = "";
        public string Title
        {
            get { return m_title; }
            set
            {
                m_title = value;
                OnPropertyChanged("Title");
            }
        }

        private string m_color = "";
        public string Color
        {
            get { return m_color; }
            set
            {
                m_color = value;
                OnPropertyChanged("Color");
            }
        }

        private string m_description = "";
        public string Description
        {
            get { return m_description; }
            set
            {
                m_description = value;
                OnPropertyChanged("Description");
            }
        }


        #region INotifyPropertyChanged Members

        public event PropertyChangedEventHandler? PropertyChanged;

        protected void OnPropertyChanged(string name)
        {
            if (PropertyChanged != null)
                PropertyChanged(this, new PropertyChangedEventArgs(name));
        }

        #endregion

        #endregion

        //TODO: 加入属性表

        public BNode(NodeKinds kind)
        {
            Kind = kind;
        }

        public IEnumerable<PortKinds> GetPorts()
        {
            switch (Kind)
            {
                case NodeKinds.Root:
                    yield return PortKinds.Right;
                    break;
                case NodeKinds.Action:
                    yield return PortKinds.Left;
                    break;
                case NodeKinds.Condition:
                    yield return PortKinds.Left;
                    yield return PortKinds.Right;
                    break;
                case NodeKinds.Composites:
                    yield return PortKinds.Left;
                    yield return PortKinds.Right;
                    break;
                case NodeKinds.Decorators:
                    yield return PortKinds.Left;
                    yield return PortKinds.Right;
                    break;

            }
        }
    }

    enum NodeKinds { Root, Action, Condition, Composites, Decorators }

    static class NodeKindConvert
    {
        public static NodeKinds ToKind(string category)
        {
            switch (category.ToLower())
            {
                case "root":
                    return NodeKinds.Root;
                case "action":
                    return NodeKinds.Action;
                case "condition":
                    return NodeKinds.Condition;
                case "composites":
                    return NodeKinds.Composites;
                case "decorators":
                    return NodeKinds.Decorators;
                default:
                    throw new ArgumentException("unknown parameters");
            }

        }
    }
}
