using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.BehaviorCharts.Model
{
    class Link : INotifyPropertyChanged
    {
        [Browsable(false)]
        public BehaviorNode Source { get; private set; }
        [Browsable(false)]
        public PortKinds SourcePort { get; private set; }
        [Browsable(false)]
        public BehaviorNode Target { get; private set; }
        [Browsable(false)]
        public PortKinds TargetPort { get; private set; }
        [Browsable(false)]

        public Link(BehaviorNode source, PortKinds sourcePort, BehaviorNode target, PortKinds targetPort)
        {
            Source = source;
            SourcePort = sourcePort;
            Target = target;
            TargetPort = targetPort;
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

    enum PortKinds { Top, Bottom, Left, Right }
}
