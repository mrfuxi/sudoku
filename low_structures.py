from collections import OrderedDict, defaultdict

import math
import cv2
import numpy as np

from visualize import draw_lines


class orderless_memoized(object):
    def __init__(self, func):
        self.func = func
        self.cache = {}

    def __call__(self, *args):
        key = frozenset(args)
        try:
            value = self.cache[key]
        except KeyError:
            value = self.func(*args)
            self.cache[key] = value
        return value


def similarly_angled_lines(line_a, line_b):
    """
    lines has to differ by some value
    otherwise intersection is not interesting
    """
    th_a = line_a[1]
    th_b = line_b[1]

    return similar_angles(th_a, th_b)


@orderless_memoized
def similar_angles(angle_a, angle_b):
    min_ang_diff = 0.5  # ~28deg

    ang_diff = abs(angle_a - angle_b)
    if ang_diff < min_ang_diff or ang_diff > (np.pi - min_ang_diff):
        return True
    return False


def intersection(line_a, line_b):
    """
    Solve:
    x*cos(th_a) + y*sin(th_a) = r_a
    x*cos(th_b) + y*sin(th_b) = r_b

    As matrix:
    A*X = b
    """
    r_a, th_a = line_a
    r_b, th_b = line_b

    A = np.array([
        [np.cos(th_a), np.sin(th_a)],
        [np.cos(th_b), np.sin(th_b)],
    ])
    b = np.array([r_a, r_b])
    ok, point = cv2.solve(A, b)
    if ok:
        point = (point[0][0], point[1][0])
    return ok, point


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


def intersections(lines):
    points = {}

    for i, line_a in enumerate(lines):
        for j, line_b in enumerate(lines):
            if i <= j:
                continue

            ok, point = intersection(line_a, line_b)
            if not ok:
                continue

            key = tuple(sorted([i, j]))
            points[key] = point

    return points


def remove_duplicate_lines(lines, img_shape, min_dist=3):
    """
    duplicates: crosses in view at low angle
    """
    w, h = img_shape
    scope = 0.5
    min_x = 0 - w*scope
    min_y = 0 - h*scope
    max_x = w + w*scope
    max_y = h + h*scope

    to_remove = set([])
    for i, line_a in enumerate(lines):
        for j, line_b in enumerate(lines[i+1:], i+1):
            similar = similarly_angled_lines(line_a, line_b)
            if not similar:
                continue

            if abs(line_a[0] - line_b[0]) < min_dist:
                to_remove.add(max(i, j))
                continue

            ok, point = intersection(line_a, line_b)
            if not ok:
                continue

            x, y = point
            in_view = (
                min_x <= x <= max_x and
                min_y <= y <= max_y
            )

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


def lines_with_similar_angle(lines, angle):
    """
    Splits list of lines into one that are similar to given angle,
    and the rest of lines
    """

    similar = []
    other = []

    for line in lines:
        if similar_angles(line[1], angle):
            similar.append(line)
        else:
            other.append(line)

    return similar, other


def put_lines_into_buckets(buckets, lines):
    bucketed = []
    all_matches = set([])
    for bucket_angle, bucket in buckets.iteritems():
        matches = []

        for line in lines:
            if is_angle_in_bucket(line[1], bucket):
                matches.append(line)

        matches_key = tuple(matches)
        if matches and matches_key not in all_matches:
            all_matches.add(matches_key)
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


def fragment_value(image, point_a, point_b):
    """
    Calculates field under a line fragment (non-zero pixels)
    Value is adjusted based on length of the fragment
    """
    w2 = 2
    rect = zip(point_a, point_b)
    start = min(rect[0]) - w2, max(rect[0]) + w2
    end = min(rect[1]) - w2, max(rect[1]) + w2
    dist = distance_between_points(point_a, point_b)

    sub_image = image[end[0]:end[1], start[0]:start[1]]

    return cv2.countNonZero(sub_image)/dist


def valid_fragment_lenghts(binded_lines):
    """
    Calculates average distance between consecutive points on the line/grid
    """
    lengths = fragment_lenghts(binded_lines)
    avg = np.mean(lengths)
    sigma = np.std(lengths)

    return (avg - sigma*2, avg + sigma*2)


def remove_very_close_lines(lines, img_shape, min_size=10):
    """
    Removes lines that are very close to each other (removes the later one)
    It should be used on lines from orthogonal buckets
    """
    binded, points = bind_intersections_to_lines(lines)
    len_min, len_max = valid_fragment_lenghts(binded)

    # lines has to be enough apart so char will fit
    # not so much apart that 9 fragments will fit
    len_min = max(len_min, min_size)
    len_max = min(len_max, min(img_shape)/9)

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
    thick_lines = draw_lines(
        image, lines, color=255, thickness=2, rgb=False, draw_on_empty=True,
    )
    thick_lines = cv2.bitwise_and(image, thick_lines)

    scores = defaultdict(list)
    values = {}

    for i, line in enumerate(binded):
        for point_a, point_b in zip(line[2], line[2][1:]):
            keys = set(points[point_a]) ^ set(points[point_b])

            score = fragment_value(thick_lines, point_a, point_b)
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
