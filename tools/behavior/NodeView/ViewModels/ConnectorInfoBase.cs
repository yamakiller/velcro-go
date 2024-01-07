using NodeBehavior.Controls;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace NodeBehavior.ViewModels
{
    public abstract class ConnectorInfoBase : INotifyPropertyChangedBase
    {
        private static double m_connectorWidth = 8;
        private static double m_connectorHeight = 8;

        public ConnectorInfoBase(ConnectorOrientation orientation)
        {
            this.Orientation = orientation;
        }

        public ConnectorOrientation Orientation { get; private set; }

        public static double ConnectorWidth
        {
            get { return m_connectorWidth; }
        }

        public static double ConnectorHeight
        {
            get { return m_connectorHeight; }
        }
    }
}
