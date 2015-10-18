from collections import OrderedDict, defaultdict
from itertools import chain

import math
import cv2
import numpy as np

from visualize import draw_lines


def similar_angle(line_a, line_b, min_ang_diff=0.5):
    """
    lines has to differ by some value
    otherwise intersection is not interesting
    """
    if min_ang_diff <= 0:
        return False

    th_a = line_a[1]
    th_b = line_b[1]

    ang_diff = np.abs(th_a - th_b)
    if ang_diff < min_ang_diff or ang_diff > (np.pi - min_ang_diff):
        return True
    return False


def intersection(line_a, line_b, min_ang_diff):
    """
    Solve:
    x*cos(th_a) + y*sin(th_a) = r_a
    x*cos(th_b) + y*sin(th_b) = r_b

    As matrix:
    A*X = b
    """
    r_a, th_a = line_a
    r_b, th_b = line_b

    if similar_angle(line_a, line_b, min_ang_diff=min_ang_diff):
        return False, None

    A = np.array([
        [np.cos(th_a), np.sin(th_a)],
        [np.cos(th_b), np.sin(th_b)],
    ])
    b = np.array([r_a, r_b])
    ok, point = cv2.solve(A, b)
    return ok, tuple(x[0] for x in point)


def point_in_view(point, img_shape, scope=0.5):
    """
    scope: padding to size. 0.5 = 50%
    """
    w, h = img_shape
    x, y = point

    min_x = 0 - w*scope
    min_y = 0 - h*scope
    max_x = w + w*scope
    max_y = h + h*scope

    return (
        min_x <= x <= max_x and
        min_y <= y <= max_y
    )


def intersections(lines, min_ang_diff=0):
    points = {}

    for i, line_a in enumerate(lines):
        for j, line_b in enumerate(lines):
            if i <= j:
                continue

            ok, point = intersection(line_a, line_b, min_ang_diff)
            if not ok:
                continue

            key = tuple(sorted([i, j]))
            points[key] = point

    return points


def remove_duplicate_lines(lines, min_ang_diff, img_shape):
    """
    duplicates: crosses in view at low angle

    min_ang_diff: angle in deg
    """
    min_ang_diff = np.deg2rad(min_ang_diff)

    to_remove = set([])
    for i, line_a in enumerate(lines):
        for j, line_b in enumerate(lines):
            if i <= j:
                continue

            similar = similar_angle(line_a, line_b, min_ang_diff)
            if not similar:
                continue

            ok, point = intersection(line_a, line_b, min_ang_diff=0)
            if not ok:
                continue

            in_view = point_in_view(point, img_shape)

            if in_view:
                to_remove.add(max(i, j))

    cleaned = [line for i, line in enumerate(lines) if i not in to_remove]
    return cleaned


def generate_angle_buckets(bucket_size, step=None, ortogonal=True):
    """
    Creates a dict with ortogonal (if required) ranges for angles (in radians)
    Both bucket_size and step are taken in deg (it's easier to reason about)
    Angles between 0 and 180 deg

    Example output (bucket_size=20, step=5) - all values in deg:
    {
        45: [(35, 55), (125, 145)],
        50: [(40, 60), (130, 150)],
    }
    """

    window = np.deg2rad(bucket_size)
    step = np.deg2rad(step or bucket_size)

    window_2 = window/2.0
    pos = 0
    max_pos = (np.pi/2.0 if ortogonal else np.pi) - step
    buckets = OrderedDict()
    while 1:
        b1 = (pos - window_2, pos + window_2)
        bucket = [b1]

        if b1[0] < 0:
            b1_prim = (np.pi + b1[0], np.pi)
            bucket.append(b1_prim)

        if b1[1] > np.pi:
            b1_bis = (0, b1[1] - np.pi)
            bucket.append(b1_bis)

        if ortogonal:
            b2 = (b1[0] + np.pi/2, b1[1] + np.pi/2)
            bucket.append(b2)

            if b2[1] > np.pi:
                b2_prim = (0, b2[1] - np.pi)
                bucket.append(b2_prim)

        buckets[pos] = bucket

        pos += step
        if pos >= max_pos:
            break

    return buckets


def is_angle_in_bucket(angle, ranges):
    """
    takes a list of ranges for a given bucket,
    and angle to test
    """

    for start, end in ranges:
        if start <= angle <= end:
            return True

    return False


def put_lines_into_buckets(buckets, lines):
    bucketed = []
    for bucket_angle, bucket in buckets.iteritems():
        matches = []

        for line in lines:
            if is_angle_in_bucket(line[1], bucket):
                matches.append(line)

        if matches:
            bucketed.append((bucket_angle, matches))

    # sort by most numbers of matches in bucket
    bucketed = sorted(bucketed, key=lambda b: len(b[1]), reverse=True)
    return bucketed


def bind_intersections_to_lines(lines):
    binded = [
        [angle, distance, []] for angle, distance in lines
    ]
    points = intersections(lines, np.deg2rad(45))
    for (line_a, line_b), point in points.items():
        binded[line_a][2].append(point)
        binded[line_b][2].append(point)

    for line in binded:
        line[2] = sorted(line[2])

    # reverse keys with values in dict
    # values (points) will be unique
    points_to_lines = {point: lines for lines, point in points.items()}

    return binded, points_to_lines


def distance_between_points(point_a, point_b):
    return math.sqrt(
        (point_a[0]-point_b[0])**2 + (point_a[1]-point_b[1])**2
    )


def fragment_lenghts(binded_lines):
    distances = []
    for distance, angle, points in binded_lines:
        for point_a, point_b in zip(points, points[1:]):
            dist = distance_between_points(point_a, point_b)
            distances.append(dist)

    return distances


def fragment_value(distance_map, point_a, point_b):
    """
    Calculates field under a line fragment (non-zero pixels)
    Value is adjusted based on length of the fragment
    """
    w2 = 2
    rect = zip(point_a, point_b)
    start = min(rect[0]) - w2, max(rect[0]) + w2
    end = min(rect[1]) - w2, max(rect[1]) + w2
    dist = distance_between_points(point_a, point_b)

    sub_image = distance_map[end[0]:end[1], start[0]:start[1]]

    return cv2.countNonZero(sub_image)/dist


def valid_fragment_lenghts(binded_lines):
    """
    Calculates average distance between consecutive points on the line/grid
    """
    lengths = fragment_lenghts(binded_lines)
    avg = np.mean(lengths)
    sigma = np.std(lengths)

    return (avg - sigma*2, avg + sigma*2)


def remove_very_close_lines(lines):
    """
    Removes lines that are very close to each other (removes the later one)
    It should be used on lines from orthogonal buckets
    """
    binded, points = bind_intersections_to_lines(lines)
    len_min, len_max = valid_fragment_lenghts(binded)

    scores = defaultdict(int)

    for i, line in enumerate(binded):
        for point_a, point_b in zip(line[2], line[2][1:]):
            dist = distance_between_points(point_a, point_b)
            score = 1
            if dist < len_min or len_max < dist:
                score = -1

            key = tuple(
                sorted(
                    set(points[point_a]) ^ set(points[point_b])
                )
            )

            scores[key] += score

    to_remove = [
        line_pair[1] for line_pair, score in scores.items() if score < 0
    ]
    return [line for i, line in enumerate(lines) if i not in to_remove]


def remove_disjonted_lines(image, lines):
    """
    Removes lines that are very close to each other (removes the later one)
    It should be used on lines from orthogonal buckets
    """
    binded, points = bind_intersections_to_lines(lines)

    # Line iterator would be better but it's not available in Python binding
    thick_lines = draw_lines(image, lines, thickness=1, rgb=False, draw_on_empty=True)
    exact_lines = draw_lines(image, lines, color=1, thickness=1, rgb=False, draw_on_empty=True)
    thick_lines = cv2.bitwise_and(image, thick_lines)
    distances = cv2.distanceTransform(thick_lines, cv2.DIST_L1, cv2.DIST_MASK_PRECISE)
    distances = distances * exact_lines

    scores = defaultdict(list)
    values = {}

    for i, line in enumerate(binded):
        for point_a, point_b in zip(line[2], line[2][1:]):
            keys = set(points[point_a]) ^ set(points[point_b])

            score = fragment_value(distances, point_a, point_b)
            for key in keys:
                scores[key].append(score)

            values[(point_a, point_b)] = score

    mean, std = np.mean(values.values()), np.std(values.values())
    to_remove = []
    for i, score in scores.items():
        if len([s for s in score if abs(s - mean) < std]) < 4:
            to_remove.append(i)

    connected = [line for i, line in enumerate(lines) if i not in to_remove]
    return connected, values
