using Behavior.Diagrams.Adorners;
using Behavior.Diagrams.Utils;
using System.Windows;
using System.Windows.Documents;
using System.Windows.Media;

namespace Behavior.Diagrams.Controls
{
    public class SegmentLink : LinkBase
    {
        static SegmentLink()
        {
            FrameworkElement.DefaultStyleKeyProperty.OverrideMetadata(
                typeof(SegmentLink), new FrameworkPropertyMetadata(typeof(LinkBase)));
        }

        public override void UpdatePath()
        {
            // 贝塞尔曲线
            var linePoints = getEndPoinds();
            if (checkPoints(linePoints))
            {
                calculatePositions(linePoints);
                PathGeometry geometry = new PathGeometry();
                PathFigure figure = new PathFigure();
                figure.StartPoint = StartPoint;
                figure.Segments.Add(new BezierSegment(MidPoint1, MidPoint2, EndPoint, true));
                geometry.Figures.Add(figure);
                this.PathGeometry = geometry;
            }
            else
            {
                this.PathGeometry = null;
            }
        }

        protected override Adorner CreateSelectionAdorner()
        {
            return new SelectionAdorner(this, new RelinkControl());
        }


        protected virtual Point[] CalculateSegments()
        {
            var res =getEndPoinds();
            if (res != null)
                UpdateEdges(res);
            return res;
        }

        /// <summary>
        /// 返回结束点
        /// </summary>
        /// <returns></returns>
        protected Point[] getEndPoinds()
        {
            Point tc, sc;
            if (Target != null)
                tc = Target.Center;
            else if (TargetPoint != null)
                tc = TargetPoint;
            else
                return null;

            if (Source != null)
                sc = Source.Center;
            else if (SourcePoint != null)
                sc = SourcePoint;
            else
                return null;

            var linePoints = new Point[2];
            linePoints[0] = sc;
            linePoints[1] = tc;
            return linePoints;
        }

        protected void UpdateEdges(Point[] linePoints)
        {
            if (linePoints.Length >= 2)
            {
                if (Source != null)
                    linePoints[0] = Source.GetEdgePoint(linePoints[1]);
                if (Target != null)
                    linePoints[linePoints.Length - 1] = Target.GetEdgePoint(linePoints[linePoints.Length - 2]);
            }
        }

        /// <summary>
        /// 计算位置
        /// </summary>
        /// <param name="linePoints"></param>
        protected virtual void calculatePositions(Point[] linePoints)
        {
            StartPoint = linePoints[0];
            EndPoint = linePoints[linePoints.Length - 1];
            StartCapAngle = GeometryHelper.NormalAngle(linePoints[0], linePoints[1]);
            EndCapAngle = GeometryHelper.NormalAngle(linePoints[linePoints.Length - 2], linePoints[linePoints.Length - 1]);

            {
                var point = GeometryHelper.SegmentMiddlePoint(StartPoint, EndPoint);
                point = GeometryHelper.SegmentMiddlePoint(StartPoint, point);
                point.Y = StartPoint.Y;

                MidPoint1 = point;
            }


            {
                var point = GeometryHelper.SegmentMiddlePoint(StartPoint, EndPoint);
                point = GeometryHelper.SegmentMiddlePoint(point, EndPoint);
                point.Y = EndPoint.Y;

                MidPoint2 = point;
            }
        }

        /// <summary>
        /// 检测点是否溢出
        /// </summary>
        /// <param name="linePoints"></param>
        /// <returns></returns>
        private bool checkPoints(Point[] linePoints)
        {
            if (linePoints != null && linePoints.Length >= 2)
            {
                foreach (var p in linePoints)
                    if (double.IsNaN(p.X) || double.IsNaN(p.Y))
                        return false;
                return true;
            }
            return false;
        }
    }
}
