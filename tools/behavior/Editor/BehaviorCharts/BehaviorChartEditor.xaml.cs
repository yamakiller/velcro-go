using Editor.BehaviorCharts.Model;
using Editor.Charts;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using System.Windows.Shapes;

namespace Editor.BehaviorCharts
{
    /// <summary>
    /// BehaviorChartEditor.xaml 的交互逻辑
    /// </summary>
    public partial class BehaviorChartEditor : UserControl
    {
        public BehaviorChartEditor()
        {
            InitializeComponent();

            var model = CreateModel();
            editor.Controller = new Controller(editor, model);
            editor.DragDropTool = new DragDropTool(editor, model);

            editor.DragTool = new CustomMoveResizeTool(editor, model)
            {
                MoveGridCell = editor.GridCellSize
            };

            editor.LinkTool = new CustomLinkTool(editor);
        }

        private BehaviorChartModel CreateModel() 
        {
            var model = new BehaviorChartModel();

            // var start = new BehaviorNode(NodeKinds.Root);
            // start.Row = 0;
            // start.Column = 1;
            // start.Name = "Root";

            return model;
        }

       

        /*void Selection_PropertyChanged(object sender, System.ComponentModel.PropertyChangedEventArgs e)
        {
            var p = editor.Selection.Primary;
            m_propertiesView.SelectedObject = p != null ? p.ModelElement : null;
        }*/
    }
}
