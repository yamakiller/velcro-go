using NodeBehavior.Controls;
using NodeBehavior.Helpers;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Net;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace NodeBehavior.ViewModels
{
    public class ConnectorViewModel : SelectableBehaviorItemViewModelBase 
    {
        private FullyCreatedConnectorInfo m_sourceConnectorInfo;
        private ConnectorInfoBase m_sinkConnectorInfo;
        private Point m_sourceB;
        private Point m_sourceA;
        private List<Point> m_connectionPoints;
        private Point m_endPoint;
        private Rect m_area;

        public Point SourceA
        {
            get
            {
                return m_sourceA;
            }
            set
            {
                if (m_sourceA != value)
                {
                    m_sourceA = value;
                    UpdateArea();
                    NotifyChanged("SourceA");
                }
            }
        }


        public Point SourceB
        {
            get
            {
                return m_sourceB;
            }
            set
            {
                if (m_sourceB != value)
                {
                    m_sourceB = value;
                    UpdateArea();
                    NotifyChanged("SourceB");
                }
            }
        }

        public List<Point> ConnectionPoints
        {
            get
            {
                return m_connectionPoints;
            }
            private set
            {
                if (m_connectionPoints != value)
                {
                    m_connectionPoints = value;
                    NotifyChanged("ConnectionPoints");
                }
            }
        }

        public Point EndPoint
        {
            get
            {
                return m_endPoint;
            }
            private set
            {
                if (m_endPoint != value)
                {
                    m_endPoint = value;
                    NotifyChanged("EndPoint");
                }
            }
        }

        public Rect Area
        {
            get
            {
                return m_area;
            }
            private set
            {
                if (m_area != value)
                {
                    m_area = value;
                    UpdateConnectionPoints();
                    NotifyChanged("Area");
                }
            }
        }

        public static IPathFinder PathFinder { get; set; }

        public bool IsFullConnection
        {
            get { return m_sinkConnectorInfo is FullyCreatedConnectorInfo; }
        }

        public ConnectorInfo ConnectorInfo(ConnectorOrientation orientation, double left, double top, Point position)
        {

            return new ConnectorInfo()
            {
                Orientation = orientation,
                DesignerItemSize = new Size(m_sourceConnectorInfo.DataItem.ItemWidth, m_sourceConnectorInfo.DataItem.ItemHeight),
                DesignerItemLeft = left,
                DesignerItemTop = top,
                Position = position

            };
        }

        public FullyCreatedConnectorInfo SourceConnectorInfo
        {
            get
            {
                return m_sourceConnectorInfo;
            }
            set
            {
                if (m_sourceConnectorInfo != value)
                {

                    m_sourceConnectorInfo = value;
                    SourceA = PointHelper.GetPointForConnector(this.SourceConnectorInfo);
                    NotifyChanged("SourceConnectorInfo");
                    (m_sourceConnectorInfo.DataItem as INotifyPropertyChanged).PropertyChanged += new WeakINotifyEventHandler(ConnectorViewModel_PropertyChanged).Handler;
                }
            }
        }



        public ConnectorInfoBase SinkConnectorInfo
        {
            get
            {
                return m_sinkConnectorInfo;
            }
            set
            {
                if (m_sinkConnectorInfo != value)
                {

                    m_sinkConnectorInfo = value;
                    if (SinkConnectorInfo is FullyCreatedConnectorInfo)
                    {
                        SourceB = PointHelper.GetPointForConnector((FullyCreatedConnectorInfo)SinkConnectorInfo);
                        (((FullyCreatedConnectorInfo)m_sinkConnectorInfo).DataItem as INotifyPropertyChanged).PropertyChanged += new WeakINotifyEventHandler(ConnectorViewModel_PropertyChanged).Handler;
                    }
                    else
                    {

                        SourceB = ((PartCreatedConnectionInfo)SinkConnectorInfo).CurrentLocation;
                    }
                    NotifyChanged("SinkConnectorInfo");
                }
            }
        }

        public ConnectorViewModel(int id, IBehaviorViewModel parent,
           FullyCreatedConnectorInfo sourceConnectorInfo, FullyCreatedConnectorInfo sinkConnectorInfo) : base(id, parent)
        {
            Init(sourceConnectorInfo, sinkConnectorInfo);
        }

        public ConnectorViewModel(FullyCreatedConnectorInfo sourceConnectorInfo, ConnectorInfoBase sinkConnectorInfo)
        {
            Init(sourceConnectorInfo, sinkConnectorInfo);
        }

        private void UpdateArea()
        {
            Area = new Rect(SourceA, SourceB);
        }

        private void UpdateConnectionPoints()
        {
            ConnectionPoints = new List<Point>()
                                   {

                                       new Point( SourceA.X  <  SourceB.X ? 0d : Area.Width, SourceA.Y  <  SourceB.Y ? 0d : Area.Height ),
                                       new Point(SourceA.X  >  SourceB.X ? 0d : Area.Width, SourceA.Y  >  SourceB.Y ? 0d : Area.Height)
                                   };

            ConnectorInfo sourceInfo = ConnectorInfo(SourceConnectorInfo.Orientation,
                                           ConnectionPoints[0].X,
                                           ConnectionPoints[0].Y,
                                           ConnectionPoints[0]);

            if (IsFullConnection)
            {
                EndPoint = ConnectionPoints.Last();
                Helpers.ConnectorInfo sinkInfo = ConnectorInfo(SinkConnectorInfo.Orientation,
                                  ConnectionPoints[1].X,
                                  ConnectionPoints[1].Y,
                                  ConnectionPoints[1]);

                ConnectionPoints = PathFinder.GetConnectionLine(sourceInfo, sinkInfo, true);
            }
            else
            {
                ConnectionPoints = PathFinder.GetConnectionLine(sourceInfo, ConnectionPoints[1], ConnectorOrientation.Left);
                EndPoint = new Point();
            }
        }

        private void ConnectorViewModel_PropertyChanged(object sender, PropertyChangedEventArgs e)
        {
            switch (e.PropertyName)
            {
                case "ItemHeight":
                case "ItemWidth":
                case "Left":
                case "Top":
                    SourceA = Helpers.PointHelper.GetPointForConnector(this.SourceConnectorInfo);
                    if (this.SinkConnectorInfo is FullyCreatedConnectorInfo)
                    {
                        SourceB = Helpers.PointHelper.GetPointForConnector((FullyCreatedConnectorInfo)this.SinkConnectorInfo);
                    }
                    break;

            }
        }

        private void Init(FullyCreatedConnectorInfo sourceConnectorInfo, ConnectorInfoBase sinkConnectorInfo)
        {
            this.Parent = sourceConnectorInfo.DataItem.Parent;
            this.SourceConnectorInfo = sourceConnectorInfo;
            this.SinkConnectorInfo = sinkConnectorInfo;
            PathFinder = new OrthogonalPathFinder();
        }
    }
}
