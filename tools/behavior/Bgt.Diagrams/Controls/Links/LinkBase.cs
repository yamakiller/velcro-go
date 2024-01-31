
using System.ComponentModel;
using System.Windows.Documents;
using System.Windows.Media;
using System.Windows;

using Bgt.Diagrams.Adorners;
using System;

namespace Bgt.Diagrams.Controls
{
    public abstract class LinkBase : DiagramItem, ILink, INotifyPropertyChanged
    {
        #region Properties

        #region CanRelink Property

        public bool CanRelink
        {
            get { return (bool)GetValue(CanRelinkProperty); }
            set { SetValue(CanRelinkProperty, value); }
        }

        public static readonly DependencyProperty CanRelinkProperty =
            DependencyProperty.Register("CanRelink",
                                       typeof(bool),
                                       typeof(LinkBase),
                                       new FrameworkPropertyMetadata(true));

        #endregion

        private IPort m_source;
        public IPort Source
        {
            get { return m_source; }
            set
            {
                if (m_source != null)
                    m_source.Links.Remove(this);
                m_source = value;
                if (m_source != null)
                    m_source.Links.Add(this);
            }
        }

        private IPort m_target;
        public IPort Target
        {
            get { return m_target; }
            set
            {
                if (m_target != null)
                    m_target.Links.Remove(this);
                m_target = value;
                if (m_target != null)
                    m_target.Links.Add(this);
            }
        }

        private IPort m_control1;
        public IPort Control1
        {
            get { return m_control1; }
            set
            {
                if (m_control1 != null)
                    m_control1.Links.Remove(this);
                m_control1 = value;
                if (m_control1 != null)
                    m_control1.Links.Add(this);
            }
        }

        private IPort m_control2;
        public IPort Control2
        {
            get { return m_control2; }
            set
            {
                if (m_control2 != null)
                    m_control2.Links.Remove(this);
                m_control2 = value;
                if (m_control2 != null)
                    m_control2.Links.Add(this);
            }
        }

        public Point SourcePoint { get; set; }
        public Point TargetPoint { get; set; }
        public Point ControlPoint1 { get; set; }
        public Point ControlPoint2 { get; set; }

        private bool m_startCap;
        public bool StartCap
        {
            get { return m_startCap; }
            set
            {
                m_startCap = value;
                OnPropertyChanged("StartCap");
            }
        }

        private bool m_endCap;
        public bool EndCap
        {
            get { return m_endCap; }
            set
            {
                m_endCap = value;
                OnPropertyChanged("EndCap");
            }
        }

        private Brush m_brush = new SolidColorBrush(Colors.Black);
        public Brush Brush
        {
            get { return m_brush; }
            set { m_brush = value; }
        }

        private Point m_startPoint;
        public Point StartPoint
        {
            get { return m_startPoint; }
            protected set
            {
                m_startPoint = value;
                OnPropertyChanged("StartPoint");
            }
        }

        private Point m_endPoint;
        public Point EndPoint
        {
            get { return m_endPoint; }
            protected set
            {
                m_endPoint = value;
                OnPropertyChanged("EndPoint");
            }
        }

        private Point m_midPoint1;
        public Point MidPoint1
        {
            get { return m_midPoint1; }
            protected set
            {
                m_midPoint1 = value;
                OnPropertyChanged("MidPoint1");
            }
        }

        private Point m_midPoint2;
        public Point MidPoint2
        {
            get { return m_midPoint2; }
            protected set
            {
                m_midPoint2 = value;
                OnPropertyChanged("MidPoint2");
            }
        }

        private double m_startCapAngle;
        public double StartCapAngle
        {
            get { return m_startCapAngle; }
            protected set
            {
                m_startCapAngle = value;
                OnPropertyChanged("StartCapAngle");
            }
        }

        private double m_endCapAngle;
        public double EndCapAngle
        {
            get { return m_endCapAngle; }
            protected set
            {
                m_endCapAngle = value;
                OnPropertyChanged("EndCapAngle");
            }
        }

        private PathGeometry m_pathGeomtry;
        public PathGeometry PathGeometry
        {
            get { return m_pathGeomtry; }
            protected set
            {
                m_pathGeomtry = value;
                OnPropertyChanged("PathGeometry");
            }
        }

        private Point m_labelPosition;
        public Point LabelPosition
        {
            get { return m_labelPosition; }
            set
            {
                m_labelPosition = value;
                OnPropertyChanged("LabelPosition");
            }
        }

        #region Label Property

        public string Label
        {
            get { return (string)GetValue(LabelProperty); }
            set { SetValue(LabelProperty, value); }
        }

        public static readonly DependencyProperty LabelProperty =
            DependencyProperty.Register("Label", typeof(string), typeof(LinkBase));

        #endregion

        public override Rect Bounds
        {
            get
            {
                var x = Math.Min(StartPoint.X, EndPoint.X);
                var y = Math.Min(StartPoint.Y, EndPoint.Y);
                var mx = Math.Max(StartPoint.X, EndPoint.X);
                var my = Math.Max(StartPoint.Y, EndPoint.Y);
                return new Rect(x, y, mx - x, my - y);
            }
        }

        #endregion

        protected LinkBase()
        {
            UpdatePath();
        }

        protected override Adorner CreateSelectionAdorner()
        {
            return new SelectionAdorner(this, new RelinkControl());
        }

        public abstract void UpdatePath();

        #region INotifyPropertyChanged Members

        public event PropertyChangedEventHandler PropertyChanged;
        protected void OnPropertyChanged(string name)
        {
            PropertyChangedEventHandler handler = PropertyChanged;
            if (handler != null)
                handler(this, new PropertyChangedEventArgs(name));
        }
        #endregion
    }
}
