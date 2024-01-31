
using System.Collections;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using Bgt.Diagrams.Controls;

namespace Bgt.Diagrams
{
    public class Selection : INotifyPropertyChanged, IEnumerable<DiagramItem>
    {
        private DiagramItem m_primary;
        public DiagramItem Primary
        {
            get { return m_primary; }
        }

        private Dictionary<DiagramItem, object> m_items = new Dictionary<DiagramItem, object>();
        public IEnumerable<DiagramItem> Items
        {
            get { return m_items.Keys; }
        }

        public int Count
        {
            get { return m_items.Count; }
        }

        internal Selection()
        {
        }

        public bool Contains(DiagramItem item)
        {
            return m_items.ContainsKey(item);
        }

        public void Add(DiagramItem item)
        {
            if (!m_items.ContainsKey(item))
            {
                bool isPrimary = Count == 0;
                m_items.Add(item, null);
                item.IsSelected = true;
                item.IsPrimarySelection = isPrimary;
                if (isPrimary)
                {
                    m_primary = item;
                    OnPropertyChanged("Primary");
                }
                OnPropertyChanged("Items");
            }
        }

        public void Remove(DiagramItem item)
        {
            if (m_items.ContainsKey(item))
            {
                item.IsSelected = false;
                m_items.Remove(item);
            }
            if (m_primary == item)
            {
                m_primary = m_items.Keys.FirstOrDefault();
                if (m_primary != null)
                    m_primary.IsPrimarySelection = true;
                OnPropertyChanged("Primary");
            }
            OnPropertyChanged("Items");
        }

        public void Set(DiagramItem item)
        {
            SetRange(new DiagramItem[] { item });
        }

        public void SetRange(IEnumerable<DiagramItem> items)
        {
            DoClear();
            bool isPrimary = true;
            foreach (var item in items)
            {
                m_items.Add(item, null);
                item.IsSelected = true;
                if (isPrimary)
                {
                    m_primary = item;
                    item.IsPrimarySelection = true;
                    isPrimary = false;
                }
            }
            OnPropertyChanged("Primary");
            OnPropertyChanged("Items");
        }

        public void Clear()
        {
            DoClear();
            OnPropertyChanged("Primary");
            OnPropertyChanged("Items");
        }

        private void DoClear()
        {
            foreach (var item in Items)
                item.IsSelected = false;
            m_items.Clear();
            m_primary = null;
        }

        #region INotifyPropertyChanged Members

        public event PropertyChangedEventHandler PropertyChanged;

        protected void OnPropertyChanged(string name)
        {
            if (PropertyChanged != null)
                PropertyChanged(this, new PropertyChangedEventArgs(name));
        }

        #endregion

        #region IEnumerable Members

        public IEnumerator GetEnumerator()
        {
            return Items.GetEnumerator();
        }

        #endregion

        #region IEnumerable<object> Members

        IEnumerator<DiagramItem> IEnumerable<DiagramItem>.GetEnumerator()
        {
            return Items.GetEnumerator();
        }

        #endregion
    }
}
