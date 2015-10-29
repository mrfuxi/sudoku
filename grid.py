#!/usr/bin/env python
import glob
from os import path, mkdir
from shutil import rmtree

import process
from low_structures import (
    distance_between_points,
    intersection,
    remove_duplicate_lines,
    put_lines_into_buckets,
    generate_angle_buckets,
    lines_with_similar_angle,
)
import visualize
import cv2

OUTDIR = 'example_out'


def point_similarities(expected_points, distances):
    fit = 0
    points = []
    step = expected_points[1] - expected_points[0]
    for expected in expected_points:
        point = distances[int(expected)]
        if points:
            fit += abs(abs(point-points[-1]) - step)
        points.append(point)

    return fit, points


def prepare_point_distances(points):
    max_p = points[-1]
    distances = [(max_p, max_p)] * int(max_p+1)
    for point in points:
        idx = int(point)
        distances[idx] = (0, point)
        for i in xrange(1, int(max_p)):
            try:
                distances[idx - i] = min((i, point), distances[idx - i])
            except IndexError:
                pass

            try:
                distances[idx + i] = min((i, point), distances[idx + i])
            except IndexError:
                pass

    return [d[1] for d in distances]


def linear_distances(lines, divider_line):
    if len(lines) < 10:
        return 0.0, []

    intersections = []
    for line in lines:
        _, point = intersection(line, divider_line, min_ang_diff=None)
        intersections.append(point)

    points = []
    for point in intersections:
        points.append(distance_between_points(intersections[0], point))

    distances = prepare_point_distances(points)

    best_fit = None
    points_count = len(points)
    res = []
    for i in range(0, points_count+1-10):
        for j in xrange(i+10, points_count+1):
            start, end = points[i], points[j-1]
            step = (end - start)/9.0
            expected_points = [start + step*k for k in range(10)]
            fit = point_similarities(expected_points, distances)

            if len(set(fit[1])) != 10:
                continue

            res.append(fit)

            best_fit = min(best_fit, fit) if best_fit is not None else fit

    if not best_fit:
        return 10**10, []

    score, best_points = best_fit
    return score, [lines[points.index(point)] for point in best_points]


def possible_grids(horizontal, vertical):
    vertical = sorted(vertical, key=lambda l: l[0])
    horizontal = sorted(horizontal, key=lambda l: l[0])

    lines_v = []
    for lines in (linear_distances(vertical, h)[1] for h in horizontal):
        if lines and lines not in lines_v:
            lines_v.append(lines)

    lines_h = []
    for lines in (linear_distances(horizontal, v)[1] for v in vertical):
        if lines and lines not in lines_h:
            lines_h.append(lines)

    grids = []
    for h in lines_h:
        for v in lines_v:
            grids.append((h, v))

    return grids


def evaluate_grids(img, grids):
    best = (0, None)
    for i, grid in enumerate(grids):
        lines_h, lines_v = grid
        fragments = []
        for h in lines_h:
            _, point_a = intersection(h, lines_v[0], min_ang_diff=None)
            _, point_b = intersection(h, lines_v[-1], min_ang_diff=None)
            fragments.append((point_a, point_b))

        for v in lines_v:
            _, point_a = intersection(v, lines_h[0], min_ang_diff=None)
            _, point_b = intersection(v, lines_h[-1], min_ang_diff=None)
            fragments.append((point_a, point_b))

        fragment_image = visualize.draw_fragments(
            img, fragments, color=255, width=2
        )
        masked_image = cv2.bitwise_and(img, fragment_image)
        score = cv2.countNonZero(masked_image)

        best = max(best, (score, grid))

    return best


def find_grid(image):
    img = process.pre_process(image)

    lines = process.find_lines(img, 100)
    dedup = remove_duplicate_lines(lines, 15, img.shape)

    bucket_size = 90.0/5.0
    buckets = generate_angle_buckets(
        bucket_size, step=bucket_size/2.0, ortogonal=True
    )
    bucketed_lines = put_lines_into_buckets(buckets, dedup)

    best = (0, None)
    for angle, line_class in bucketed_lines:
        # don't even bother doing any more work
        # it's not a 9x9 grid
        if len(line_class) < 20:
            continue

        vertical, horizontal = lines_with_similar_angle(line_class, angle, 0.5)

        if len(vertical) < 10 or len(horizontal) < 10:
            continue

        grids = possible_grids(horizontal, vertical)
        score, grid = evaluate_grids(img, grids)
        best = max(best, (score, grid))

    return best[0], best[1], line_class

if __name__ == '__main__':
    rmtree(OUTDIR, ignore_errors=True)
    mkdir(OUTDIR)

    for filename in sorted(glob.glob('examples/*.png')):
        filename = path.basename(filename)
        print filename
        img = process.get_example_image(filename)
        score, grid, all_lines = find_grid(img)

        if grid:
            result = visualize.draw_lines(img, grid[0] + grid[1], thickness=2)
            cv2.imwrite("{}/{}".format(OUTDIR, filename), result)
        else:
            print "No grid found"
