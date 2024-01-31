using Bgt.Diagrams.Controls;
using Bgt.Diagrams.Utils;
using Bgt.Diagrams;

using Editor.BehaviorCharts.Model;
using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Shapes;

namespace Editor.BehaviorCharts
{
    class Controller : IDiagramController
    {
        private class UpdateScope : IDisposable
        {
            private Controller m_parent;
            public bool IsInprogress { get; set; }

            public UpdateScope(Controller parent)
            {
                m_parent = parent;
            }

            public void Dispose()
            {
                IsInprogress = false;
                m_parent.UpdateView();
            }
        }

        private DiagramView m_view;
        private EditorViewModel m_model;
        private UpdateScope m_updateScope;

        public Controller(DiagramView view, EditorViewModel model)
        {
            m_view = view;
            m_model = model;
            m_model.Nodes.CollectionChanged += NodesCollectionChanged;
            m_model.Links.CollectionChanged += LinksCollectionChanged;
            m_updateScope = new UpdateScope(this);

            foreach (var t in m_model.Nodes)
                t.PropertyChanged += NodePropertyChanged;

            UpdateView();
        }

        void NodesCollectionChanged(object sender, 
                                    System.Collections.Specialized.NotifyCollectionChangedEventArgs e)
        {
            if (e.NewItems != null)
                foreach (var t in e.NewItems.OfType<INotifyPropertyChanged>())
                    t.PropertyChanged += NodePropertyChanged;

            if (e.OldItems != null)
                foreach (var t in e.OldItems.OfType<INotifyPropertyChanged>())
                    t.PropertyChanged -= NodePropertyChanged;
            UpdateView();
        }

        void LinksCollectionChanged(object sender, 
            System.Collections.Specialized.NotifyCollectionChangedEventArgs e)
        {
            UpdateView();
        }

        void NodePropertyChanged(object sender, PropertyChangedEventArgs e)
        {
            var fn = sender as BehaviorNode;
            var n = m_view.Children.OfType<Node>().FirstOrDefault(p => p.ModelElement == fn);
            if (fn != null && n != null)
                UpdateNode(fn, n);
        }

        private void UpdateView()
        {
            if (!m_updateScope.IsInprogress)
            {
                m_view.Children.Clear();

                foreach (var node in m_model.Nodes)
                    m_view.Children.Add(UpdateNode(node, null));

                foreach (var link in m_model.Links)
                    m_view.Children.Add(CreateLink(link));
            }
        }

        private Node UpdateNode(BehaviorNode node, Node item)
        {
            if (item == null)
            {
                item = new Node();
                item.ModelElement = node;
                CreatePorts(node, item);
                item.Content = CreateContent(node);
            }

            item.Width = 50;
            item.Height = 50;
            item.CanResize = false;

            item.SetValue(Canvas.LeftProperty, node.Column * m_view.GridCellSize.Width + 10);
            item.SetValue(Canvas.TopProperty, node.Row * m_view.GridCellSize.Height + 25);
            return item;
        }

        public static FrameworkElement CreateContent(BehaviorNode node)
        {
            var textBlock = new TextBlock()
            {
                VerticalAlignment = VerticalAlignment.Center,
                HorizontalAlignment = HorizontalAlignment.Center
            };

            var b = new Binding("Text");
            b.Source = node;
            textBlock.SetBinding(TextBlock.TextProperty, b);
            if (node.Kind == NodeKinds.Root)
            {
                var ui = new Ellipse();
                ui.Width = ui.Height = 50;
                ui.Stroke = Brushes.Black;
                ui.Fill = Brushes.Yellow;

                var grid = new Grid();
                grid.Children.Add(ui);

                node.Text = "Root";

                grid.Children.Add(textBlock);
                return grid;
            }
            else if (node.Kind == NodeKinds.Action)
            {
                var ui = new Border();
                ui.BorderBrush = Brushes.Black;
                ui.BorderThickness = new Thickness(1);
                ui.Background = Brushes.Lime; ;
                ui.Child = textBlock;
                return ui;
            }
            else
            {
                var ui = new Path();
                ui.Stroke = Brushes.Black;
                ui.StrokeThickness = 1;
                ui.Fill = Brushes.Pink;
                var converter = new GeometryConverter();
                // TODO: 根据类型更换图标
                ui.Data = (Geometry)converter.ConvertFrom("M 0,0.25 L 0.5 0 L 1,0.25 L 0.5,0.5 Z");
                ui.Stretch = Stretch.Uniform;

                var grid = new Grid();
                grid.Children.Add(ui);
                grid.Children.Add(textBlock);

                return grid;
            }
        }

        private void CreatePorts(BehaviorNode node, Node item)
        {
            foreach (var kind in node.GetPorts())
            {
                var port = new EllipsePort();
                port.Width = 6;
                port.Height = 6;
                port.Margin = new Thickness(-5);
                port.Visibility = Visibility.Visible;
                port.VerticalAlignment = ToVerticalAligment(kind);
                port.HorizontalAlignment = ToHorizontalAligment(kind);
                port.CanAcceptIncomingLinks = true /*(kind == PortKinds.Top)*/;
                port.CanAcceptOutgoingLinks = true /*!port.CanAcceptIncomingLinks*/;
                port.Tag = kind;
                port.Cursor = Cursors.Cross;
                port.CanCreateLink = true;
                item.Ports.Add(port);
            }
        }

        private Control CreateLink(Link link)
        {
            var item = new OrthogonalLink();
            item.ModelElement = link;
            item.EndCap = true;
            item.Source = FindPort(link.Source, link.SourcePort);
            item.Target = FindPort(link.Target, link.TargetPort);

            var b = new Binding("Name");
            b.Source = link;
            item.SetBinding(LinkBase.LabelProperty, b);

            return item;
        }

        private IPort FindPort(BehaviorNode node, PortKinds portKind)
        {
            var inode = m_view.Items.FirstOrDefault(p => p.ModelElement == node) as INode;
            if (inode == null)
                return null;
            var port = inode.Ports.OfType<FrameworkElement>().FirstOrDefault(
                p => p.VerticalAlignment == ToVerticalAligment(portKind)
                    && p.HorizontalAlignment == ToHorizontalAligment(portKind)
                );
            return (IPort)port;
        }

        private VerticalAlignment ToVerticalAligment(PortKinds kind)
        {
            if (kind == PortKinds.Top)
                return VerticalAlignment.Top;
            if (kind == PortKinds.Bottom)
                return VerticalAlignment.Bottom;
            else
                return VerticalAlignment.Center;
        }

        private HorizontalAlignment ToHorizontalAligment(PortKinds kind)
        {
            if (kind == PortKinds.Left)
                return HorizontalAlignment.Left;
            if (kind == PortKinds.Right)
                return HorizontalAlignment.Right;
            else
                return HorizontalAlignment.Center;
        }

        private void DeleteSelection()
        {
            using (BeginUpdate())
            {
                var nodes = m_view.Selection.Select(p => p.ModelElement as BehaviorNode).Where(p => p != null);
                var links = m_view.Selection.Select(p => p.ModelElement as Link).Where(p => p != null);
                foreach(var p in m_model.Nodes)
                {
                    if (nodes.Contains(p))
                    {
                        m_model.Nodes.Remove(p);
                    }
                }
                foreach (var p in m_model.Links)
                {
                    if (links.Contains(p))
                    {
                        m_model.Links.Remove(p);
                    }
                }

                foreach (var p in m_model.Links)
                {
                    if (nodes.Contains(p.Source) || nodes.Contains(p.Target))
                    {
                        m_model.Links.Remove(p);
                    }
                }
            }
        }

        private IDisposable BeginUpdate()
        {
            m_updateScope.IsInprogress = true;
            return m_updateScope;
        }

        #region IDiagramController Members

        public void UpdateItemsBounds(DiagramItem[] items, Rect[] bounds)
        {
            for (int i = 0; i < items.Length; i++)
            {
                var node = items[i].ModelElement as BehaviorNode;
                if (node != null)
                {
                    node.Column = (int)(bounds[i].X / m_view.GridCellSize.Width);
                    node.Row = (int)(bounds[i].Y / m_view.GridCellSize.Height);
                }
            }
        }

        public void UpdateLink(LinkInfo initialState, ILink link)
        {
            using (BeginUpdate())
            {
                var sourcePort = link.Source as PortBase;
                var source = VisualHelper.FindParent<Node>(sourcePort);
                var targetPort = link.Target as PortBase;
                var target = VisualHelper.FindParent<Node>(targetPort);

                m_model.Links.Remove((link as LinkBase).ModelElement as Link);
                m_model.Links.Add(
                    new Link(
                        (BehaviorNode)source.ModelElement, (PortKinds)sourcePort.Tag,
                        (BehaviorNode)target.ModelElement, (PortKinds)targetPort.Tag
                        ));
            }
        }

        public void ExecuteCommand(ICommand command, object parameter)
        {
            if (command == ApplicationCommands.Delete)
                DeleteSelection();
        }

        public bool CanExecuteCommand(ICommand command, object parameter)
        {
            if (command == ApplicationCommands.Delete)
                return true;
            else
                return false;
        }

        #endregion
    }
}
