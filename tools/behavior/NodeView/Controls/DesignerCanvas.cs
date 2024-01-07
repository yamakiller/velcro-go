using NodeBehavior.ViewModels;
using System;
using System.Collections.Generic;
using System.Drawing;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Controls;

namespace NodeBehavior.Controls
{
    public class DesignerCanvas : Canvas
    {
        private ConnectorViewModel m_partialConnection;
        private List<Connector> m_connectorsHits = new List<Connector>();
        private Connector m_sourceConnector;
        private Point? m_rubberbandSelectionStartPoint = null;

        public Connector SourceConnector
        {
            get { return m_sourceConnector; }
            set
            {
                if (m_sourceConnector != value)
                {
                    m_sourceConnector = value;
                    m_connectorsHits.Add(m_sourceConnector);
                    FullyCreatedConnectorInfo sourceDataItem = m_sourceConnector.DataContext as FullyCreatedConnectorInfo;
                }
            }
        }
    }
}
