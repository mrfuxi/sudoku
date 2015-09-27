from collections import OrderedDict

import cv2
import numpy as np

def similar_angle(line_a, line_b, min_ang_diff=0.5):
    """
    lines has to differ by some value
    otherwise intersection is not interesting
    """
    if min_ang_diff <= 0:
        return False

    th_a = line_a[0][1]
    th_b = line_b[0][1]

    ang_diff = np.abs(th_a - th_b)
    if ang_diff < min_ang_diff or ang_diff > (np.pi - min_ang_diff):
        return True
    return False


def intersection(line_a, line_b, min_ang_diff=0):
    """
    Solve:
    x*cos(th_a) + y*sin(th_a) = r_a
    x*cos(th_b) + y*sin(th_b) = r_b

    As matrix:
    A*X = b
    """
    r_a, th_a = line_a[0]
    r_b, th_b = line_b[0]

    # if min_ang_diff and similar_angle(line_a, line_b, min_ang_diff):
    #     return None, (0, 0)

    A = np.array([
        [np.cos(th_a), np.sin(th_a)],
        [np.cos(th_b), np.sin(th_b)],
    ])
    b = np.array([r_a, r_b])
    ret, dst = cv2.solve(A, b)
    return ret, dst


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

            ok, point = intersection(line_a, line_b)
            if not ok:
                continue

            in_view = point_in_view(point, img_shape)

            if in_view:
                to_remove.add(max(i, j))

    cleaned = [line for i, line in enumerate(lines) if i not in to_remove]
    return cleaned


def generate_angle_buckets(bucket_size, step, ortogonal=True):
    """
    Creates a dict with ortogonal (if required) ranges for angles (in radians)
    Both bucket_size and step are taken in deg (to make it easier to reason about)
    Angles between 0 and 180 deg

    Example output (bucket_size=20, step=5) - all values in deg:
    {
        45: [(35, 55), (125, 145)],
        50: [(40, 60), (130, 150)],
    }
    """

    step = np.deg2rad(step)
    window = np.deg2rad(bucket_size)

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
