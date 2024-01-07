using NodeBehavior.Controls;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NodeBehavior.ViewModels
{
    public class FullyCreatedConnectorInfo : ConnectorInfoBase
    {
        private bool showConnectors = false;

        public BehaviorItemViewModelBase DataItem { get; private set; }

        public bool ShowConnectors
        {
            get
            {
                return showConnectors;
            }
            set
            {
                if (showConnectors != value)
                {
                    showConnectors = value;
                    NotifyChanged("ShowConnectors");
                }
            }
        }

        public FullyCreatedConnectorInfo(BehaviorItemViewModelBase dataItem, ConnectorOrientation orientation)
            : base(orientation)
        {
            this.DataItem = dataItem;
        }
    }
}
