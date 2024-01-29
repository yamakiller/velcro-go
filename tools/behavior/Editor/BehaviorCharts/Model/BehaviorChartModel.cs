
using System.Collections.ObjectModel;


namespace Editor.BehaviorCharts.Model
{
    class BehaviorChartModel
    {
        private ObservableCollection<BehaviorNode> m_nodes = new ObservableCollection<BehaviorNode> ();
        internal ObservableCollection<BehaviorNode> Nodes
        { 
            get { return m_nodes; } 
        }

        private ObservableCollection<Link> m_links = new ObservableCollection<Link> ();
        internal ObservableCollection<Link> Links
        {
            get { return this.m_links; }
        }
    }
}
