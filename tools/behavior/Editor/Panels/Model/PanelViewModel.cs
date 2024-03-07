﻿using Behavior.Diagrams;

using Editor.Datas;
using Editor.Framework;

using Editor.Utils;
using Editor.ViewModels;
using System;
using System.Collections.Generic;
using System.Collections.ObjectModel;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Input;
using System.Windows.Media;
using System.Xml;

namespace Editor.Panels.Model
{
    class PanelViewModel : PaneViewModel
    {
        #region 常量
        private static int UnitColumnGap = 4;
        private static int UnitRow = 3;
        private static int UnitRowCap = 2;
        #endregion

        #region 成员
        private EditorFrameViewModel? m_parentViewModel = null;
        private DiagramView? m_editor = null;
        private BehaviorTree? m_btree = null;
        #endregion

        #region 节点/线表
        private ObservableCollection<BNode> m_nodes = new ObservableCollection<BNode>();
        internal ObservableCollection<BNode> Nodes
        {
            get { return m_nodes; }
        }

        private ObservableCollection<BLink> m_links = new ObservableCollection<BLink>();
        internal ObservableCollection<BLink> Links
        {
            get { return m_links; }
        }
        #endregion

        #region 指令

        #region loaded 指令
        private PaneCommand? m_loadedCommand;
        public ICommand LoadedCommand
        {
            get
            {
                if (m_loadedCommand == null)
                {
                    m_loadedCommand = new PaneCommand((p) => onLoaded(p), (p) => isLoaded(p));
                }
                return m_loadedCommand;
            }
        }
        #endregion

        #region close 指令
        private PaneCommand? m_closeCommand;
        public ICommand CloseCommand
        {
            get
            {
                if (m_closeCommand == null)
                {
                    m_closeCommand = new PaneCommand((p) => onClose(p), (p) => isClose());
                }
                return m_closeCommand;
            }
        }
        #endregion

        #region insert node 指令
        private PaneCommand? m_insertCommand;
        public PaneCommand InsertCommand
        {
            get
            {
                if (m_insertCommand == null)
                {
                    m_insertCommand = new PaneCommand((p) => onInsert(p));
                }
                return m_insertCommand;
            }
        }
        #endregion

        #endregion

        #region 构造
        public PanelViewModel(EditorFrameViewModel viewModel, BehaviorTree t)
        {
            m_parentViewModel = viewModel;
            m_btree = t;

            ContentId = t.ID;
            Title = t.Title;
        }
        #endregion

        #region 私有方法
        private bool isLoaded(object parameter)
        {
            return true;
        }

        private bool isClose()
        {
            return true;
        }

        private void onLoaded(object parameter)
        {
            DiagramView editor = parameter as DiagramView;
            Debug.Assert(editor != null);
            Debug.Assert(m_btree != null);
            Debug.Assert(m_btree.Nodes != null);

            m_editor = editor;
            if (m_btree.Nodes != null)
            {
                Datas.BehaviorNode? rootNode = null;
                foreach (var node in m_btree.Nodes)
                {
                    var kind = NodeKindConvert.ToKind(node.Value.Category);
                    if (kind == NodeKinds.Root)
                    {
                        rootNode = node.Value;
                        break;
                    }
                }

                if (rootNode != null)
                {
                    var rootBNode = new BNode(NodeKinds.Root);
                    rootBNode.Id = rootNode.ID;
                    rootBNode.Row = 30;
                    rootBNode.Column = 1;
                    rootBNode.Color = rootNode.Color;// "#FFB8860B";

                    Nodes.Add(rootBNode);

                    if (rootNode.Children != null)
                    {
                        int childCount = rootNode.Children.Count;
                        int childIndex = 0;
                        foreach (var child in rootNode.Children)
                        {
                            var childNode = m_btree.Nodes.GetValueOrDefault(child, null);
                            if (childNode == null) continue;
                            constructNode(rootBNode, childNode, childCount, childIndex);
                            childIndex++;
                        }
                    }
                }
            }

            editor.Controller = new PanelController(editor, this);
            editor.DragDropTool = new PanelDragDropTool(editor, this);

            editor.Selection.PropertyChanged += new System.ComponentModel.PropertyChangedEventHandler(onSelectionPropertyChanged);
        }

        /// <summary>
        /// 关闭处理函数
        /// </summary>
        private void onClose(object parameter)
        {
            var parentBNode = parameter as BNode;
            Debug.Assert(m_parentViewModel != null);
            Debug.Assert(m_btree != null);


            DelNode(parentBNode);

            //m_parentViewModel.CloseBehaviorTreeView(m_btree);
        }

        private void onInsert(object parameter)
        {
            var parentBNode = parameter as BNode;
            Debug.Assert(parentBNode != null);
            if (parentBNode.Kind == Model.NodeKinds.Root)
            {
                NewNode(parentBNode, "Selector", "composites");
            }
            else
            {
                NewNode(parentBNode, "Runner", "action");
            }
        }




        /// <summary>
        /// 构建节点
        /// </summary>
        /// <param name="parentBNode"></param>
        /// <param name="node"></param>
        /// <param name="count"></param>
        /// <param name="index"></param>
        void constructNode(BNode parentBNode, Datas.BehaviorNode node, int count, int index)
        {
            Debug.Assert(m_btree != null);
            Debug.Assert(m_btree.Nodes != null);

            var nodeKind = NodeKindConvert.ToKind(node.Category);
            var newBNode = new BNode(nodeKind);
            newBNode.Id = node.ID;
            newBNode.Name = node.Name;
            newBNode.Color = node.Color;
            newBNode.Description = node.Description;
            newBNode.Category = node.Category;
            newBNode.Title = node.Title;

            if (string.IsNullOrEmpty(newBNode.Color))
            {
                switch (nodeKind)
                {
                    case NodeKinds.Condition:
                        newBNode.Color = "#FFDEB887";
                        break;
                    case NodeKinds.Decorators:
                        newBNode.Color = "#FFBDB76B";
                        break;
                    case NodeKinds.Composites:
                        newBNode.Color = "#FF87CEEB";
                        break;
                    default:
                        newBNode.Color = "#FF00FF7F";
                        break;
                }
            }

            newBNode.Column = parentBNode.Column + (parentBNode.Width / 20) + UnitColumnGap;

            if (count == 1)
            {
                newBNode.Row = parentBNode.Row;
            }
            else
            {
                int h = (count * UnitRow) + ((count - 1) * UnitRowCap);
                int startRow = (parentBNode.Row - (h / 2)) + 1;
                newBNode.Row = startRow + ((index * UnitRow) + (index * UnitRowCap));
            }


            Nodes.Add(newBNode);

            if (parentBNode != null) // 增加连接线
            {
                Links.Add(new BLink(parentBNode, Model.PortKinds.Right, newBNode, Model.PortKinds.Left));
            }

            if (node.Children == null) { return; }
            int childCount = node.Children.Count;
            int childIndex = 0;
            foreach (var child in node.Children)
            {
                var childNode = m_btree.Nodes.GetValueOrDefault(child, null);
                if (childNode == null) continue;
                constructNode(newBNode, childNode, childCount, childIndex);
                childIndex++;
            }
        }
        #endregion

        #region 公共方法
        /// <summary>
        /// 创建一个新节点
        /// </summary>
        /// <param name="parent"></param>
        /// <param name="name"></param>
        /// <param name="category"></param>
        public void NewNode(BNode parent,
                            string name,
                            string category)
        {

            Debug.Assert(parent != null);
            Debug.Assert(m_parentViewModel != null);
            Debug.Assert(m_btree != null);
            Datas.BehaviorNode? parentNodeData = FindNode(parent.Id);
            if (parentNodeData == null) { return; }
            if (parent.Kind == Model.NodeKinds.Root &&
                parentNodeData.Children != null &&
                parentNodeData.Children.Count > 0)
            {
                return;
            }

            var nodeKind = NodeKindConvert.ToKind(category);

            BehaviorNode newNode = new BehaviorNode(m_parentViewModel)
            {
                ID = ShortGuid.Next(),
                Name = name,
                Category = category,
                Title = "",
                Description = "",
            };


            switch (nodeKind)
            {
                case NodeKinds.Condition:
                    newNode.Color = "#FFDEB887";
                    break;
                case NodeKinds.Decorators:
                    newNode.Color = "#FFBDB76B";
                    break;
                case NodeKinds.Composites:
                    newNode.Color = "#FF87CEEB";
                    break;
                default:
                    newNode.Color = "#FF00FF7F";
                    break;
            }


            if (m_btree.Nodes == null)
                m_btree.Nodes = new Dictionary<string, BehaviorNode>();

            int insertIndex = Nodes.Count;
            m_btree.Nodes.Add(newNode.ID, newNode);
            if (parentNodeData.Children == null)
                parentNodeData.Children = new ObservableCollection<string>();

            if (parentNodeData.Children != null && parentNodeData.Children.Count >= 1)
            {
                int idx = FindeBNodeIndex(parentNodeData.Children[parentNodeData.Children.Count - 1]);
                if (idx >= 0)
                {
                    insertIndex = idx + 1;
                }
            }
            else
            {
                int idx = FindeBNodeIndex(parentNodeData.ID);
                if (idx >= 0)
                {
                    insertIndex = idx + 1;
                }
            }


            parentNodeData.Children?.Add(newNode.ID);

            // TODO: 在视图中显示节点

            var newBNode = new BNode(nodeKind);
            newBNode.Id = newNode.ID;
            newBNode.Name = newNode.Name;
            newBNode.Category = newNode.Category;
            newBNode.Title = newNode.Title;
            newBNode.Color = newNode.Color;
            newBNode.Description = newNode.Description;
            
            if (insertIndex >= Nodes.Count)
            {
                Nodes.Add(newBNode);
            }
            else
            {
                Nodes.Insert(insertIndex, newBNode);
            }





            newBNode.Column = parent.Column + (parent.Width / 20) + UnitColumnGap;

            if (parentNodeData.Children != null)
            {
                if (parentNodeData.Children.Count == 1)
                {
                    newBNode.Row = parent.Row;
                }
                else
                {
                    int h = (parentNodeData.Children.Count * UnitRow) + ((parentNodeData.Children.Count - 1) * UnitRowCap);
                    int startRow = (parent.Row - (h / 2)) + 1;
                    foreach (var curr in parentNodeData.Children)
                    {
                        var currNode = FindBNode(curr);
                        if (currNode == null)
                        {
                            continue;
                        }

                        currNode.Row = startRow;
                        startRow += UnitRow + UnitRowCap;
                    }
                }
            }



            if (parent != null) // 增加连接线
            {
                Links.Add(new BLink(parent, Model.PortKinds.Right, newBNode, Model.PortKinds.Left));
            }
        }

        /// <summary>
        /// 删除一个新节点
        /// </summary>
        /// <param name="parent"></param>
        public void DelNode(BNode parent)
        {
            if (parent == null) return;
            BLink? source = FindSourceBlink(parent);
            while (source != null)
            {
                DelNode(source.Target);
                source = FindSourceBlink(source.Source);
            }

            BLink? target = FindTargetBlink(parent);
            if (target != null)
            {
                Datas.BehaviorNode? sourceNodeData = FindNode(target.Source.Id);
                if (sourceNodeData != null)
                {
                    for (int j = 0; j < sourceNodeData.Children.Count; j++)
                    {
                        if (sourceNodeData.Children[j] == parent.Id)
                        {
                            sourceNodeData.Children.RemoveAt(j);
                            break;
                        }
                    }
                    for (int j = 0; j < sourceNodeData.Children.Count; j++)
                    {

                    }
                        if (sourceNodeData.Children.Count == 1)
                    {
                    }
                    else
                    {
                        int h = (sourceNodeData.Children.Count * UnitRow) + ((sourceNodeData.Children.Count - 1) * UnitRowCap);
                        int startRow = (parent.Row - (h / 2)) + 1;
                        foreach (var curr in sourceNodeData.Children)
                        {
                            var currNode = FindBNode(curr);
                            if (currNode == null)
                            {
                                continue;
                            }

                            currNode.Row = startRow;
                            startRow += UnitRow + UnitRowCap;
                        }
                    }
                }
                Links.Remove(target);
            }

            Nodes.Remove(parent);
            m_btree.Nodes.Remove(parent.Id);
        }

        public void ResetSize()
        {

        }

        /// <summary>
        /// 通过ID在图中查找节点
        /// </summary>
        /// <param name="id"></param>
        /// <returns></returns>
        public BehaviorNode? FindNode(string id)
        {
            Debug.Assert(m_btree != null);
            if (m_btree.Nodes != null)
            {
                foreach (var node in m_btree.Nodes)
                {
                    if (node.Value.ID == id)
                    {
                        return node.Value;
                    }
                }
            }

            return null;
        }

        private BNode? FindBNode(string id)
        {
            foreach (var curr in Nodes)
            {
                if (curr.Id == id)
                {
                    return curr;
                }
            }

            return null;
        }

        private int FindeBNodeIndex(string id)
        {
            for (int i = 0; i < Nodes.Count; i++)
            {
                if (Nodes[i].Id == id)
                {
                    return i;
                }
            }
            return -1;
        }

        private BLink? FindSourceBlink(BNode source)
        {
            for (int lc = 0; lc < Links.Count; lc++)
            {
                if (Links[lc].Source == source)
                {
                    return Links[lc];
                }
            }
            return null;
        }

        private BLink? FindTargetBlink(BNode target)
        {
            for (int lc = 0; lc < Links.Count; lc++)
            {
                if (Links[lc].Target == target)
                {
                    return Links[lc];
                }
            }
            return null;
        }
        #endregion

        #region 事件
        void onSelectionPropertyChanged(object sender, System.ComponentModel.PropertyChangedEventArgs e)
        {
            Debug.Assert(m_editor != null);
            Debug.Assert(m_parentViewModel != null);
            var p = m_editor.Selection.Primary;
            m_parentViewModel.PropertiesSelectedObject = p != null ? p.ModelElement : null;
        }
        #endregion
    }
}
