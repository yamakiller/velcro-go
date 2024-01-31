using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.BehaviorCharts.Model
{
    class BehaviorNode : INotifyPropertyChanged
    {
        [Browsable(false)]
        public NodeKinds Kind { get; private set; }

        private int m_column;
        public int Column { get { return m_column; } set { m_column = value; OnPropertyChanged("Column"); } }
        private int m_row;
        public int Row { get { return m_row; } set { m_row = value; OnPropertyChanged("Row"); } }

        private string m_text;
        public string Text
        {
            get { return m_text; }
            set
            {
                m_text = value;
                OnPropertyChanged("Text");
            }
        }

        //TODO: 加入属性表

        public BehaviorNode(NodeKinds kind)
        {
            Kind = kind;
        }

        public IEnumerable<PortKinds> GetPorts()
        {
            switch (Kind)
            {
                case NodeKinds.Root:
                // 					yield return PortKinds.Bottom;
                // 					break;
    
                case NodeKinds.Action:
                // 					yield return PortKinds.Top;
                // 					yield return PortKinds.Bottom;
                // 					break;
                case NodeKinds.Condition:
                    yield return PortKinds.Top;
                    yield return PortKinds.Bottom;
                    yield return PortKinds.Left;
                    yield return PortKinds.Right;
                    break;
                case NodeKinds.Composites:
                    yield return PortKinds.Top;
                    yield return PortKinds.Bottom;
                    yield return PortKinds.Left;
                    yield return PortKinds.Right;
                    break;
                case NodeKinds.Decorators:
                    yield return PortKinds.Top;
                    yield return PortKinds.Bottom;
                    yield return PortKinds.Left;
                    yield return PortKinds.Right;
                    break;

            }
        }


        #region INotifyPropertyChanged Members

        public event PropertyChangedEventHandler PropertyChanged;

        protected void OnPropertyChanged(string name)
        {
            if (PropertyChanged != null)
                PropertyChanged(this, new PropertyChangedEventArgs(name));
        }

        #endregion
    }


    enum NodeKinds { Root,  Action, Condition, Composites, Decorators }
}
