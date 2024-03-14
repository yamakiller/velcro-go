using Behavior.Diagrams;
using Behavior.Diagrams.Controls;
using Behavior.Diagrams.Utils;
using Editor.Contrels;
using Editor.Converters;
using Editor.Panels.Model;
using System.ComponentModel;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Media3D;
using System.Windows.Shapes;
using System.Xml.Linq;

namespace Editor.Panels
{
    class PanelController : IDiagramController
    {
        private class UpdateScope : IDisposable
        {
            private PanelController m_parent;
            public bool IsInprogress { get; set; }

            public UpdateScope(PanelController parent)
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
        private PanelViewModel m_model;
        private UpdateScope m_updateScope;
        public PanelController(DiagramView view, PanelViewModel model)
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
            var fn = sender as BNode;
            var n = m_view.Children.OfType<Behavior.Diagrams.Controls.Node>().FirstOrDefault(p => p.ModelElement == fn);
            if (fn != null && n != null)
                UpdateNode(fn, n,e);
        }


        private void UpdateView()
        {
            if (!m_updateScope.IsInprogress)
            {
                m_view.Children.Clear();

                foreach (var node in m_model.Nodes)
                    m_view.Children.Add(UpdateNode(node, null,null));

                m_model.ResetNodeRow();


                foreach (var link in m_model.Links)
                    m_view.Children.Add(CreateLink(link));
            }
        }

        private Behavior.Diagrams.Controls.Node UpdateNode(BNode node, Behavior.Diagrams.Controls.Node? item, PropertyChangedEventArgs e)
        {
            if (item == null)
            {
                item = new Behavior.Diagrams.Controls.Node();
                item.ModelElement = node;
                CreatePorts(node, item);
                item.Content = CreateContent(node);
            }
            if(e?.PropertyName == "Category")
            {
                if (NodeKindConvert.ToKind(node.Category) == NodeKinds.Action)
                {
                    m_model.RemoveAllChildNode(node);
                }
                node.Color = NodeKindConvert.ToColor(NodeKindConvert.ToKind(node.Category));
                item.Content = CreateContent(node);
            }

            item.Width = node.Width;
            item.Height = node.Hegith;
            item.IsResize = false;

            item.SetValue(Canvas.LeftProperty, node.Column * m_view.GridCellSize.Width + 10);
            item.SetValue(Canvas.TopProperty, node.Row * m_view.GridCellSize.Height + 25);

            return item;
        }

        public FrameworkElement CreateContent(BNode node)
        {
            if (node.Kind == Model.NodeKinds.Root)
            {
                //node.Name = "Root";
                return CreateNode(node,
                                  160,
                                  new SolidColorBrush(ColorHelper.ToColor(node.Color)),
                                  Geometry.Parse("M513.6 417.1c41.5-23 83.2-45.6 124.6-69 14.6-8.3 26.5-7.4 38.8 4.6 17.4 16.9 37.6 29.7 60.5 38.1 112.8 41.2 235.6-54.1 223.7-173.5-9.9-99-99.7-169.2-196.2-153.4C673 79 611.4 167.2 627.7 260.7c4.4 25.4 3.1 28.1-19.6 40.6-39.2 21.7-78.4 43.3-117.6 64.9-21.2 11.7-26.3 10.9-42.7-6.8-62.5-67.5-159.4-89.9-245-56.6-84.6 32.9-140.4 115-140.4 206.6 0 101.9 67.3 190.5 165.9 215.2 78.8 19.8 148.8 0.2 209-54.1 11.1-10 21.1-15.5 35.3-6.2 46.8 30.8 93.8 61.2 140.7 91.7 13.8 9 18.4 21.2 15.1 37.1-2.7 13-4.7 26.3-3.4 39.7 7.5 80.1 72.3 134.5 154.6 127.5 52.4-4.5 92-31.5 113.3-79.7 20.9-47.3 19-95.1-11.7-138.4-45.3-63.9-124.9-80.1-190.9-37.5-21.6 13.9-36.9 12.8-57.2-1.5-41.8-29.5-85.2-56.8-128.5-84.1-14.3-9-18-18.6-11.7-34.4 17-43 18.8-87.1 6.1-131.7-6.4-21.6-4.7-25.2 14.6-35.9z m280.1-298c62.4 0.5 111.6 50.1 111.5 112.2-0.1 62.5-49.2 111.4-111.9 111.4-63.2 0.1-113-49.7-112.6-112.6 0.4-61.6 51.2-111.5 113-111z m-28.1 618.6c46 0.8 84.7 40.2 84 85.6-0.7 45.1-39.6 82.8-85 82.4-46.5-0.4-84.5-39-83.8-85.1 0.7-44.9 40.4-83.7 84.8-82.9z"),
                                  0x1);
            }
            else if (node.Kind == Model.NodeKinds.Condition) 
            {
                return CreateNode(node,
                                  260,
                                  new SolidColorBrush(ColorHelper.ToColor(node.Color)),
                                  Geometry.Parse("M304.6 487.8c-13.4 0-24.2 10.9-24.2 24.2 0 13.4 10.9 24.2 24.2 24.2 13.4 0 24.2-10.9 24.2-24.2 0.1-13.4-10.8-24.2-24.2-24.2z m0-138.3c-13.4 0-24.2 10.9-24.2 24.2 0 13.4 10.9 24.2 24.2 24.2 6.5 0 12.6-2.5 17.1-7.1 4.6-4.6 7.1-10.7 7.1-17.1 0.1-13.3-10.8-24.2-24.2-24.2zM280.4 650.2a24.2 24.2 0 1 0 48.4 0 24.2 24.2 0 1 0-48.4 0z\r\nM512 1024c-14.2 0-28.2-3.8-40.5-10.9l-384-224.3c-24.6-14.4-39.8-40.8-39.8-69.4V304.2c0-28.4 15.2-55 39.8-69.4l384-224.3c24-14 57.1-14 81.1 0l383.9 224.3c24.6 14.4 39.8 40.8 39.8 69.4v415.2c-0.1 28.6-15.2 54.9-39.8 69.4l-384 224.2c-12.3 7.2-26.3 11-40.5 11z m0-1001.6c-10.5 0-20.6 2.6-29.1 7.6L99 254.4c-17.7 10.4-28.5 29.3-28.6 49.8v415.2c0 20.4 11 39.5 28.6 49.8l384 224.3c17.6 10.3 40.5 10.3 58.2 0L925 769.2c17.7-10.4 28.6-29.3 28.6-49.8V304.2c0-20.4-11-39.5-28.6-49.8L541.1 30.1c-8.5-5-18.6-7.7-29.1-7.7z m0 0\r\nM728 630.5H417c-6.1 0-11.1 5-11.1 11.1v17.3c0 6.1 5 11.1 11.1 11.1h311c6.1 0 11.1-5 11.1-11.1v-17.3c0-6.1-5-11.1-11.1-11.1z m0-138.2H417c-6.1 0-11.1 5-11.1 11.1v17.3c0 6.1 5 11.1 11.1 11.1h311c6.1 0 11.1-5 11.1-11.1v-17.3c0-6.1-5-11.1-11.1-11.1z m0-138.2H417c-6.1 0-11.1 5-11.1 11.1v17.3c0 6.1 5 11.1 11.1 11.1h311c6.1 0 11.1-5 11.1-11.1v-17.3c0-6.2-5-11.1-11.1-11.1z"),
                                  0x03);
            }
            else if (node.Kind == Model.NodeKinds.Decorators)
            {
                return CreateNode(node,
                                  260,
                                  new SolidColorBrush(ColorHelper.ToColor(node.Color)),
                                  Geometry.Parse("M891.908919 710.299325V313.862667a15.8195 15.8195 0 0 0-10.915454-15.8195l-363.848489-132.251015a17.085059 17.085059 0 0 0-11.231845 0L142.381032 298.359557a15.8195 15.8195 0 0 0-10.915454 15.8195V710.299325a16.768669 16.768669 0 0 0 10.915454 15.819499l363.848489 132.092821a17.876034 17.876034 0 0 0 5.69502 0.94917 15.8195 15.8195 0 0 0 5.536825-0.94917l363.848489-132.092821a15.8195 15.8195 0 0 0 10.599064-15.819499z m-396.436657 82.577787v25.943979L196.958306 710.299325 495.472262 601.144778v26.102174h33.062753V601.144778l298.513956 109.154547-298.513956 108.521766v-25.943979z m33.062753-226.377038v-37.966799h-33.062753v37.966799L165.161112 686.570075V337.433721l330.31115 120.070002v37.966798h33.062753v-37.966798l330.31115-120.070002V686.570075z m-15.819499-367.170583l314.966235 114.533176-315.599015 114.533176-315.12443-114.533176z\r\nM1024.00174 512.080996A512.551784 512.551784 0 0 0 512.082736 0.003797a506.223984 506.223984 0 0 0-300.570491 97.448117 81.786812 81.786812 0 0 0-47.458498-14.87033 82.577787 82.577787 0 0 0-82.577788 82.577788 81.312227 81.312227 0 0 0 14.87033 47.458498A506.223984 506.223984 0 0 0 0.005537 512.080996a511.760809 511.760809 0 0 0 811.856715 414.629082 82.577787 82.577787 0 0 0 114.849566-114.849566 507.647739 507.647739 0 0 0 97.289922-299.779516zM165.161112 115.644338a49.515033 49.515033 0 1 1-49.515034 49.515034 49.515033 49.515033 0 0 1 49.515034-49.515034zM33.068291 512.080996a474.584985 474.584985 0 0 1 88.589197-276.999437 82.261397 82.261397 0 0 0 113.425812-113.425811 474.584985 474.584985 0 0 1 276.999436-88.589197 479.48903 479.48903 0 0 1 478.85625 479.014445A473.003035 473.003035 0 0 1 901.717009 788.922237 82.261397 82.261397 0 0 0 788.923977 901.715269a473.003035 473.003035 0 0 1-276.841241 88.589197A479.48903 479.48903 0 0 1 33.068291 512.080996z m825.777874 396.278463a49.515033 49.515033 0 1 1 49.515034-49.515034 49.515033 49.515033 0 0 1-49.515034 49.515034z\r\nM495.472262 660.626096h33.062753v33.062754h-33.062753zM495.472262 726.751604h33.062753V759.339773h-33.062753z"),
                                  0x03);
            }
            else if (node.Kind == Model.NodeKinds.Composites)
            {
                return CreateNode(node,
                                  260,
                                  new SolidColorBrush(ColorHelper.ToColor(node.Color)),
                                  Geometry.Parse("M66.56 473.6c-5.12 5.12-10.24 12.8-10.24 23.04 0 7.68 2.56 15.36 10.24 23.04l158.72 163.84c5.12 2.56 12.8 7.68 23.04 7.68 7.68 0 17.92-2.56 23.04-10.24l166.4-163.84c5.12-5.12 10.24-12.8 10.24-23.04s-2.56-17.92-10.24-23.04l-166.4-161.28c-5.12-5.12-12.8-10.24-23.04-10.24-7.68 0-15.36 2.56-23.04 10.24L66.56 473.6zM967.68 81.92c0-7.68-2.56-15.36-10.24-23.04-5.12-5.12-15.36-10.24-23.04-10.24H680.96c-7.68 0-15.36 2.56-23.04 10.24-5.12 5.12-10.24 15.36-10.24 23.04v64H504.32c-7.68 0-15.36 2.56-23.04 10.24-5.12 5.12-10.24 15.36-10.24 23.04v640c0 7.68 2.56 15.36 10.24 23.04 5.12 5.12 15.36 10.24 23.04 10.24h143.36v94.72c0 7.68 2.56 15.36 10.24 23.04 5.12 5.12 15.36 10.24 23.04 10.24H934.4c7.68 0 15.36-2.56 23.04-10.24 5.12-5.12 10.24-15.36 10.24-23.04V739.84c0-7.68-2.56-15.36-10.24-23.04-5.12-5.12-15.36-10.24-23.04-10.24H680.96c-7.68 0-15.36 2.56-23.04 10.24-5.12 5.12-10.24 15.36-10.24 23.04v15.36h-64c-7.68 0-15.36-2.56-23.04-10.24-5.12-5.12-10.24-15.36-10.24-23.04V256c0-7.68 2.56-15.36 10.24-23.04 5.12-5.12 15.36-10.24 23.04-10.24h64v64c0 7.68 2.56 15.36 10.24 23.04 5.12 5.12 15.36 10.24 23.04 10.24H934.4c7.68 0 15.36-2.56 23.04-10.24 5.12-5.12 10.24-15.36 10.24-23.04V81.92z"),
                                  0x03);
            }
            else
            {
                return CreateNode(node,
                                  260,
                                  new SolidColorBrush(ColorHelper.ToColor(node.Color)),
                                  Geometry.Parse("M0.102605 0.665267v1023.283559h1024.460569V0.665267H0.102605zM922.163174 921.651174H102.553779V102.962919H922.163174V921.651174z m-273.57841-415.894452L358.681715 665.77911V358.783808l289.903049 146.972914z"),
                                  0x02);
            }
        }

        private void CreatePorts(BNode node, Behavior.Diagrams.Controls.Node item)
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

                port.Tag = kind;
                port.Cursor = Cursors.Cross;
               

                item.Ports.Add(port);
            }
        }

        private Control CreateLink(BLink link)
        {
            var item = new OrthogonalLink();
            item.ModelElement = link;
            item.EndCap = true;
            item.Source = FindPort(link.Source, link.SourcePort);
            item.Target = FindPort(link.Target, link.TargetPort);

 
            return item;
        }

        private Grid CreateNode(BNode node,
                                double width,
                                SolidColorBrush brushColor,
                                Geometry icon,
                                int btn)
        {

            var textBlock = new TextBlock()
            {
                VerticalAlignment = VerticalAlignment.Center,
                HorizontalAlignment = HorizontalAlignment.Center
            };

            var b = new Binding("Name");
            b.Source = node;
            textBlock.SetBinding(TextBlock.TextProperty, b);



            var blackui = new Rectangle();
            blackui.Width = 160;
            blackui.Height = 50;
            blackui.RadiusX = 2;
            blackui.RadiusY = 2;
            blackui.Stroke = brushColor;
            blackui.StrokeThickness = 3;
            blackui.Fill = Brushes.White;
           

            var grid = new Grid();
            grid.Children.Add(blackui);


            var capui = new Graphics.RoundedCornersPolygon
            {
                StrokeThickness = 1,
                Fill = brushColor,
                ArcRoundness = 2,
                UseRoundnessPercentage = false,
                IsClosed = true
            };

            var colorbing = new Binding("Color");
            colorbing.Source = node;
            colorbing.Converter = StringToSolidColorBrushConverter.Instance;
            capui.SetBinding(Graphics.RoundedCornersPolygon.FillProperty, colorbing);

            double innerCircle = blackui.StrokeThickness + 1;
            capui.Points.Add(new Point(innerCircle, innerCircle));
            capui.Points.Add(new Point(blackui.Width - innerCircle, innerCircle));
            capui.Points.Add(new Point(blackui.Width - innerCircle, 46));
            capui.Points.Add(new Point(innerCircle, 46));

            grid.Children.Add(capui);

            var capicon = new Path();
            capicon.Stroke = Brushes.Black;
            capicon.StrokeThickness = 1;
            capicon.Fill = Brushes.Black;
            capicon.Stretch = Stretch.Uniform;
            capicon.Data = icon;
            capicon.Width = 16;
            capicon.Height = 16;

            textBlock.FontSize = 16;
            textBlock.Foreground = Brushes.Black;
            textBlock.Margin = new Thickness(4, 0, 0, 0);

            var stackPanel = new StackPanel();
            stackPanel.HorizontalAlignment = HorizontalAlignment.Center;
            stackPanel.Orientation = Orientation.Horizontal;

            stackPanel.Children.Add(capicon);
            stackPanel.Children.Add(textBlock);

            if ((btn & 0x01) != 0)
            {
                AddButton addBtn = new AddButton();
                addBtn.Margin = new Thickness(4, 0, 0, 0);
                addBtn.Width = 16;
                addBtn.Height = 16;
                // 加入点击事件
                addBtn.Command = m_model.InsertCommand;
                addBtn.CommandParameter = node;

                stackPanel.Children.Add(addBtn);
            }

            if ((btn & 0x02) != 0)
            {
                DelButton delBtn = new DelButton();
                delBtn.Margin = new Thickness(4, 0, 0, 0);
                delBtn.Width = 16;
                delBtn.Height = 16;

                // 加入点击事件
                delBtn.Command = m_model.CloseCommand;
                delBtn.CommandParameter = node;

                stackPanel.Children.Add(delBtn);
            }

            grid.Children.Add(stackPanel);


            node.Width = (int)blackui.Width;
            node.Hegith = (int)blackui.Height;

            return grid;
        }

        private IPort? FindPort(BNode node, Model.PortKinds portKind)
        {
            var inode = m_view.Items.FirstOrDefault(p => p.ModelElement == node) as INode;
            if (inode == null) return null;
            var port = inode.Ports.OfType<FrameworkElement>().FirstOrDefault(
                p => p.VerticalAlignment == ToVerticalAligment(portKind)
                    && p.HorizontalAlignment == ToHorizontalAligment(portKind)
                );
            if (port == null) return null;
            return (IPort)port;
        }

        private VerticalAlignment ToVerticalAligment(Model.PortKinds kind)
        {
            if (kind == Model.PortKinds.Top)
                return VerticalAlignment.Top;
            if (kind == Model.PortKinds.Bottom)
                return VerticalAlignment.Bottom;
            else
                return VerticalAlignment.Center;
        }

        private HorizontalAlignment ToHorizontalAligment(Model.PortKinds kind)
        {
            if (kind == Model.PortKinds.Left)
                return HorizontalAlignment.Left;
            if (kind == Model.PortKinds.Right)
                return HorizontalAlignment.Right;
            else
                return HorizontalAlignment.Center;
        }

        private void DeleteSelection()
        {
            using (BeginUpdate())
            {
                var nodes = m_view.Selection.Select(p => p.ModelElement as Model.BNode).Where(p => p != null);
                var links = m_view.Selection.Select(p => p.ModelElement as Model.BLink).Where(p => p != null);
                foreach (var p in m_model.Nodes)
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
                var node = items[i].ModelElement as BNode;
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
                var source = VisualHelper.FindParent<Behavior.Diagrams.Controls.Node>(sourcePort);
                var targetPort = link.Target as PortBase;
                var target = VisualHelper.FindParent<Behavior.Diagrams.Controls.Node>(targetPort);
                var pl = (link as LinkBase)?.ModelElement as BLink;

                if (pl != null)
                {
                    m_model.Links.Remove(pl);
                }

                if (sourcePort != null && targetPort != null)
                {
                    m_model.Links.Add(
                        new BLink(
                            (BNode)source.ModelElement, (Model.PortKinds)sourcePort.Tag,
                            (BNode)target.ModelElement, (Model.PortKinds)targetPort.Tag
                            ));
                }
            }
        }
        #endregion
    }
}
