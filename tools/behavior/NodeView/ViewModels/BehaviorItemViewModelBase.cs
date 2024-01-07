using NodeBehavior.Controls;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NodeBehavior.ViewModels
{
    //https://github.com/sachabarber/MVVMDiagramDesigner/blob/master/DiagramDesignerMVVM/DiagramDesigner/Controls/DesignerCanvas.cs
    public abstract class BehaviorItemViewModelBase : SelectableBehaviorItemViewModelBase
    {
        private double m_left;
        private double m_top;
        
        private double m_itemWidth = 65;
        private double m_itemHeight = 65;

        private bool m_showConnectors = false;
        private List<FullyCreatedConnectorInfo> m_connectors = new List<FullyCreatedConnectorInfo>();
        public double Left
        {
            get
            {
                return m_left;
            }
            set
            {
                if (m_left != value)
                {
                    m_left = value;
                    NotifyChanged("Left");
                }
            }
        }

        public double Top
        {
            get
            {
                return m_top;
            }
            set
            {
                if (m_top != value)
                {
                    m_top = value;
                    NotifyChanged("Top");
                }
            }
        }


        public BehaviorItemViewModelBase(int id, IBehaviorViewModel parent, double left, double top) : base(id, parent)
        {
            this.m_left = left;
            this.m_top = top;
            Init();
        }

        public BehaviorItemViewModelBase(int id, IBehaviorViewModel parent, double left, double top, double itemWidth, double itemHeight) : base(id, parent)
        {
            this.m_left = left;
            this.m_top = top;
            this.m_itemWidth = itemWidth;
            this.m_itemHeight = itemHeight;
            Init();
        }

        public BehaviorItemViewModelBase() : base()
        {
            Init();
        }

        public double ItemWidth
        {
            get
            {
                return m_itemWidth;
            }
            set
            {
                if (m_itemWidth != value)
                {
                    m_itemWidth = value;
                    NotifyChanged("ItemWidth");
                }
            }
        }

        public double ItemHeight
        {
            get
            {
                return m_itemHeight;
            }
            set
            {
                if (m_itemHeight != value)
                {
                    m_itemHeight = value;
                    NotifyChanged("ItemHeight");
                }
            }
        }

        public FullyCreatedConnectorInfo TopConnector
        {
            get { return m_connectors[0]; }
        }


        public FullyCreatedConnectorInfo BottomConnector
        {
            get { return m_connectors[1]; }
        }


        public FullyCreatedConnectorInfo LeftConnector
        {
            get { return m_connectors[2]; }
        }


        public FullyCreatedConnectorInfo RightConnector
        {
            get { return m_connectors[3]; }
        }


        public bool ShowConnectors
        {
            get
            {
                return m_showConnectors;
            }
            set
            {
                if (m_showConnectors != value)
                {
                    m_showConnectors = value;
                    TopConnector.ShowConnectors = value;
                    BottomConnector.ShowConnectors = value;
                    RightConnector.ShowConnectors = value;
                    LeftConnector.ShowConnectors = value;
                    NotifyChanged("ShowConnectors");
                }
            }
        }


        private void Init()
        {
            m_connectors.Add(new FullyCreatedConnectorInfo(this, ConnectorOrientation.Top));
            m_connectors.Add(new FullyCreatedConnectorInfo(this, ConnectorOrientation.Bottom));
            m_connectors.Add(new FullyCreatedConnectorInfo(this, ConnectorOrientation.Left));
            m_connectors.Add(new FullyCreatedConnectorInfo(this, ConnectorOrientation.Right));
        }
    }
}
