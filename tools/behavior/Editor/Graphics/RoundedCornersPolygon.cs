using System;
using System.Windows;
using System.Windows.Media;
using System.Windows.Shapes;

namespace Editor.Graphics
{
    public class RoundedCornersPolygon : Shape
    {
        private readonly Path m_path;

        #region Properties 
        private PointCollection m_points;
        public PointCollection Points
        {
            get { return m_points; }
            set
            {
                m_points = value;
                RedrawShape();
            }
        }

        private bool m_isClosed;
        public bool IsClosed
        {
            get
            {
                return m_isClosed;
            }
            set
            {
                m_isClosed = value;
                RedrawShape();
            }
        }


        private bool m_useRoundnessPercentage;
        /// <summary>
        /// 获取或设置一个值，该值指定 ArcRoundness 属性值是否将用作连接线段的百分比.
        /// </summary>
        public bool UseRoundnessPercentage
        {
            get
            {
                return m_useRoundnessPercentage;
            }
            set
            {
                m_useRoundnessPercentage = value;
                RedrawShape();
            }
        }

        private double m_arcRoundness;
        /// <summary>
        /// 获取或设置指定圆弧圆度的值.
        /// </summary>
        public double ArcRoundness
        {
            get
            {
                return m_arcRoundness;
            }
            set
            {
                m_arcRoundness = value;
                RedrawShape();
            }
        }

        public Geometry Data
        {
            get
            {
                return m_path.Data;
            }
        }
        #endregion

        public RoundedCornersPolygon()
        {
            var geometry = new PathGeometry();
            geometry.Figures.Add(new PathFigure());
            m_path = new System.Windows.Shapes.Path { Data = geometry };
            Points = new PointCollection();
            Points.Changed += Points_Changed;
        }

        private void Points_Changed(object sender, EventArgs e)
        {
            RedrawShape();
        }

        #region Implementation of Shape
        protected override Geometry DefiningGeometry
        {
            get
            {
                return m_path.Data;
            }
        }
        #endregion

        #region Private Methods
        private void RedrawShape()
        {
            var pathGeometry = m_path.Data as PathGeometry;
            if (pathGeometry == null) return;

            var pathFigure = pathGeometry.Figures[0];

            pathFigure.Segments.Clear();

            for (int counter = 0; counter < Points.Count; counter++)
            {
                switch (counter)
                {
                    case 0:
                        AddPointToPath(Points[counter], null, null);
                        break;
                    case 1:
                        AddPointToPath(Points[counter], Points[counter - 1], null);
                        break;
                    default:
                        AddPointToPath(Points[counter], Points[counter - 1], Points[counter - 2]);
                        break;
                }
            }

            if (IsClosed)
                CloseFigure(pathFigure);
        }

        /// <summary>
        /// Adds a point to the shape
        /// </summary>
        /// <param name="currentPoint">The current point added</param>
        /// <param name="prevPoint">Previous point</param>
        /// <param name="prevPrevPoint">The point before the previous point</param>
        private void AddPointToPath(Point currentPoint, Point? prevPoint, Point? prevPrevPoint)
        {
            if (Points.Count == 0)
                return;

            var pathGeometry = m_path.Data as PathGeometry;
            if (pathGeometry == null) return;

            var pathFigure = pathGeometry.Figures[0];

            //the first point of a polygon
            if (prevPoint == null)
            {
                pathFigure.StartPoint = currentPoint;
            }
            //second point of the polygon, only a line will be drawn
            else if (prevPrevPoint == null)
            {
                var lines = new LineSegment { Point = currentPoint };
                pathFigure.Segments.Add(lines);
            }
            //third point and above
            else
            {
                ConnectLinePoints(pathFigure, prevPrevPoint.Value, prevPoint.Value, currentPoint, ArcRoundness, UseRoundnessPercentage);
            }
        }

        /// <summary>
        /// Adds the segments necessary to close the shape
        /// </summary>
        /// <param name="pathFigure"></param>
        private void CloseFigure(PathFigure pathFigure)
        {
            //No need to visually close the figure if we don't have at least 3 points.
            if (Points.Count < 3)
                return;
            Point backPoint, nextPoint;
            if (UseRoundnessPercentage)
            {
                backPoint = GetPointAtDistancePercent(Points[Points.Count - 1], Points[0], ArcRoundness, false);
                nextPoint = GetPointAtDistancePercent(Points[0], Points[1], ArcRoundness, true);
            }
            else
            {
                backPoint = GetPointAtDistance(Points[Points.Count - 1], Points[0], ArcRoundness, false);
                nextPoint = GetPointAtDistance(Points[0], Points[1], ArcRoundness, true);
            }
            ConnectLinePoints(pathFigure, Points[Points.Count - 2], Points[Points.Count - 1], backPoint, ArcRoundness, UseRoundnessPercentage);
            var line2 = new QuadraticBezierSegment { Point1 = Points[0], Point2 = nextPoint };
            pathFigure.Segments.Add(line2);
            pathFigure.StartPoint = nextPoint;
        }

        /// <summary>
        /// Method used to connect 2 segments with a common point, defined by 3 points and aplying an arc segment between them
        /// </summary>
        /// <param name="pathFigure"></param>
        /// <param name="p1">First point, of the first segment</param>
        /// <param name="p2">Second point, the common point</param>
        /// <param name="p3">Third point, the second point of the second segment</param>
        /// <param name="roundness">The roundness of the arc</param>
        /// <param name="usePercentage">A value that indicates if the roundness of the arc will be used as a percentage or not</param>
        private static void ConnectLinePoints(PathFigure pathFigure, Point p1, Point p2, Point p3, double roundness, bool usePercentage)
        {
            //The point on the first segment where the curve will start.
            Point backPoint;
            //The point on the second segment where the curve will end.
            Point nextPoint;
            if (usePercentage)
            {
                backPoint = GetPointAtDistancePercent(p1, p2, roundness, false);
                nextPoint = GetPointAtDistancePercent(p2, p3, roundness, true);
            }
            else
            {
                backPoint = GetPointAtDistance(p1, p2, roundness, false);
                nextPoint = GetPointAtDistance(p2, p3, roundness, true);
            }

            int lastSegmentIndex = pathFigure.Segments.Count - 1;
            //Set the ending point of the first segment.
            ((LineSegment)(pathFigure.Segments[lastSegmentIndex])).Point = backPoint;
            //Create and add the curve.
            var curve = new QuadraticBezierSegment { Point1 = p2, Point2 = nextPoint };
            pathFigure.Segments.Add(curve);
            //Create and add the new segment.
            var line = new LineSegment { Point = p3 };
            pathFigure.Segments.Add(line);
        }

        /// <summary>
        /// Gets a point on a segment, defined by two points, at a given distance.
        /// </summary>
        /// <param name="p1">First point of the segment</param>
        /// <param name="p2">Second point of the segment</param>
        /// <param name="distancePercent">Distance percent to the point</param>
        /// <param name="firstPoint">A value that indicates if the distance is calculated by the first or the second point</param>
        /// <returns></returns>
        private static Point GetPointAtDistancePercent(Point p1, Point p2, double distancePercent, bool firstPoint)
        {
            double rap = firstPoint ? distancePercent / 100 : (100 - distancePercent) / 100;
            return new Point(p1.X + (rap * (p2.X - p1.X)), p1.Y + (rap * (p2.Y - p1.Y)));
        }

        /// <summary>
        /// Gets a point on a segment, defined by two points, at a given distance.
        /// </summary>
        /// <param name="p1">First point of the segment</param>
        /// <param name="p2">Second point of the segment</param>
        /// <param name="distance">Distance  to the point</param>
        /// <param name="firstPoint">A value that indicates if the distance is calculated by the first or the second point</param>
        /// <returns>The point calculated.</returns>
        private static Point GetPointAtDistance(Point p1, Point p2, double distance, bool firstPoint)
        {
            double segmentLength = Math.Sqrt(Math.Pow((p2.X - p1.X), 2) + Math.Pow((p2.Y - p1.Y), 2));
            //The distance cannot be greater than half of the length of the segment
            if (distance > (segmentLength / 2))
                distance = segmentLength / 2;
            double rap = firstPoint ? distance / segmentLength : (segmentLength - distance) / segmentLength;
            return new Point(p1.X + (rap * (p2.X - p1.X)), p1.Y + (rap * (p2.Y - p1.Y)));
        }
        #endregion
    }
}
