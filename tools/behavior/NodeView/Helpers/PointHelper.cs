using NodeBehavior.Controls;
using NodeBehavior.ViewModels;
using System.Windows;

namespace NodeBehavior.Helpers
{
    public class PointHelper
    {
        public static Point GetPointForConnector(FullyCreatedConnectorInfo connector)
        {
            Point point = new Point();

            switch (connector.Orientation)
            {
                case ConnectorOrientation.Top:
                    point = new Point(connector.DataItem.Left + (connector.DataItem.ItemWidth / 2), connector.DataItem.Top - (ConnectorInfoBase.ConnectorHeight));
                    break;
                case ConnectorOrientation.Bottom:
                    point = new Point(connector.DataItem.Left + (connector.DataItem.ItemWidth / 2), (connector.DataItem.Top + connector.DataItem.ItemHeight) + (ConnectorInfoBase.ConnectorHeight / 2));
                    break;
                case ConnectorOrientation.Right:
                    point = new Point(connector.DataItem.Left + connector.DataItem.ItemWidth + (ConnectorInfoBase.ConnectorWidth), connector.DataItem.Top + (connector.DataItem.ItemHeight / 2));
                    break;
                case ConnectorOrientation.Left:
                    point = new Point(connector.DataItem.Left - ConnectorInfoBase.ConnectorWidth, connector.DataItem.Top + (connector.DataItem.ItemHeight / 2));
                    break;
            }

            return point;
        }
    }
}
