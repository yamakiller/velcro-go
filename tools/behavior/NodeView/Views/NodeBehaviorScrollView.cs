using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Controls;
using System.Windows.Input;
using System.Windows.Threading;

namespace NodeBehavior.Views
{
    public class NodeBehaviorScrollView : ScrollViewer
    {
        double m_dx = 0;
        double m_dy = 0;
        private DispatcherTimer m_timer;

        public double Sensitivity { get; set; }
        public double ScrollStep { get; set; }
        public double Delay
        {
            get { return m_timer.Interval.TotalMilliseconds; }
            set { m_timer.Interval = TimeSpan.FromMilliseconds(value); }
        }

        public NodeBehaviorScrollView()
        {
            m_timer = new DispatcherTimer();
            m_timer.Tick += new EventHandler(Tick);

            HorizontalScrollBarVisibility = ScrollBarVisibility.Auto;
            VerticalScrollBarVisibility = ScrollBarVisibility.Auto;
            Focusable = false;
            Sensitivity = 20;
            ScrollStep = 16;
            Delay = 50;
        }

        private void Tick(object sender, EventArgs e)
        {
            if (!(Content is NodeBehaviorView) || !((NodeBehaviorView)Content).IsDragging)
                return;

            if (m_dx != 0)
                this.ScrollToHorizontalOffset(this.HorizontalOffset + m_dx);
            if (m_dy != 0)
                this.ScrollToVerticalOffset(this.VerticalOffset + m_dy);
        }

        protected override void OnPreviewMouseMove(MouseEventArgs e)
        {
            if (!(Content is NodeBehaviorView) || !((NodeBehaviorView)Content).IsDragging)
            {
                m_timer.IsEnabled = false;
            }
            else
            {
                m_timer.IsEnabled = true;
                var point = e.GetPosition(this);
                m_dx = m_dy = 0;
                if (point.X < Sensitivity)
                    m_dx = -ScrollStep;
                else if (point.X > this.ActualWidth - Sensitivity)
                    m_dx = +ScrollStep;

                if (point.Y < Sensitivity)
                    m_dy = -ScrollStep;
                else if (point.Y > this.ActualHeight - Sensitivity)
                    m_dy = +ScrollStep;
            }
            base.OnPreviewMouseMove(e);
        }
    }
}
