using Bga.Diagrams.Utils;
using System.Windows;
using System.Windows.Media;

namespace Bga.Diagrams.Controls
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
            /*var linePoints = CalculateSegments();
            if (CheckPoints(linePoints))
            {
                CalculatePositions(linePoints);
                PathGeometry geometry = new PathGeometry();
                PathFigure figure = new PathFigure();
                figure.StartPoint = linePoints[0];
                figure.Segments.Add(new PolyLineSegment(linePoints.Skip(1), true));
                geometry.Figures.Add(figure);
                this.PathGeometry = geometry;
            }
            else
                this.PathGeometry = null;*/
            // 贝塞尔曲线
            var linePoints = GetEndPoinds();
            if (CheckPoints(linePoints))
            {
                CalculatePositions(linePoints);
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

        protected virtual Point[] CalculateSegments()
        {
            var res = GetEndPoinds();
            if (res != null)
                UpdateEdges(res);
            return res;
        }

        protected Point[] GetEndPoinds()
        {
            Point tc, sc;
            if (Target != null)
                tc = Target.Center;
            else if (TargetPoint != null)
                tc = TargetPoint.Value;
            else
                return null;

            if (Source != null)
                sc = Source.Center;
            else if (SourcePoint != null)
                sc = SourcePoint.Value;
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

        protected virtual void CalculatePositions(Point[] linePoints)
        {
            StartPoint = linePoints[0];
            EndPoint = linePoints[linePoints.Length - 1];
            StartCapAngle = GeometryHelper.NormalAngle(linePoints[0], linePoints[1]);
            EndCapAngle = GeometryHelper.NormalAngle(linePoints[linePoints.Length - 2], linePoints[linePoints.Length - 1]);
            
            if (ControlPoint1 == null)
            {
                var point = GeometryHelper.SegmentMiddlePoint(StartPoint, EndPoint);
                point = GeometryHelper.SegmentMiddlePoint(StartPoint, point);
                point.Y -= 50;
                MidPoint1 = point;
            }
            else
            {
                MidPoint1 = ControlPoint1.Value;
            }

            if (ControlPoint2 == null)
            {
                var point = GeometryHelper.SegmentMiddlePoint(StartPoint, EndPoint);
                point = GeometryHelper.SegmentMiddlePoint(point, EndPoint);
                point.Y -= 50;
                MidPoint2 = point;
            }
            else
            {
                MidPoint2 = ControlPoint2.Value;
            }
            
            var mid = (int)(linePoints.Length / 2);
            var p = GeometryHelper.SegmentMiddlePoint(linePoints[mid - 1], linePoints[mid]);
            LabelPosition = new Point(p.X, p.Y - 15);
        }

        private bool CheckPoints(Point[] linePoints)
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
