#!/usr/bin/env python
import glob
from collections import defaultdict
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
            f = abs(abs(point-points[-1]) - step)
            if f >= 0.2 * step:
                break
            fit += f

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
        yield 0.0, []

    intersections = []
    for line in lines:
        _, point = intersection(line, divider_line, min_ang_diff=None)
        intersections.append(point)

    points = []
    for point in intersections:
        points.append(distance_between_points(intersections[0], point))

    distances = prepare_point_distances(points)

    points_count = len(points)
    for i in range(0, points_count+1-10):
        for j in xrange(i+10, points_count+1):
            start, end = points[i], points[j-1]
            step = (end - start)/9.0
            expected_points = [start + step*k for k in range(10)]
            fit = point_similarities(expected_points, distances)

            if len(set(fit[1])) != 10:
                continue

            score, selected_points = fit
            yield score, [lines[points.index(point)] for point in selected_points]


def possible_grids(horizontal, vertical):
    vertical = sorted(vertical, key=lambda l: l[0])
    horizontal = sorted(horizontal, key=lambda l: l[0])

    lines_v = defaultdict(float)
    for fit in (linear_distances(vertical, h) for h in horizontal):
        for score, line in fit:
            lines_v[tuple(line)] += score

    lines_h = defaultdict(float)
    for fit in (linear_distances(horizontal, v) for v in vertical):
        for score, line in fit:
            lines_h[tuple(line)] += score

    lines_v = [l for l, s in sorted(lines_v.items(), key=lambda x: x[1])]
    lines_h = [l for l, s in sorted(lines_h.items(), key=lambda x: x[1])]

    grids = []
    for h in lines_h[:3]:
        for v in lines_v[:3]:
            grids.append((h, v))

    return grids


def evaluate_grids(img, grids):
    best = (10**10, None)
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
        score = masked_image.sum()

        # max dark ink, so minimize it
        best = min(best, (score, grid))

    return best


def find_grid(image):
    img = process.pre_process(image)
    img_grey = process.gray_image(image)

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
        score, grid = evaluate_grids(img_grey, grids)
        best = max(best, (score, grid))

    return best

if __name__ == '__main__':
    rmtree(OUTDIR, ignore_errors=True)
    mkdir(OUTDIR)

    for filename in sorted(glob.glob('examples/*.png')):
        filename = path.basename(filename)
        print filename
        img = process.get_example_image(filename)
        score, grid = find_grid(img)

        if grid:
            result = visualize.draw_lines(img, grid[0] + grid[1], thickness=2)
            cv2.imwrite("{}/{}".format(OUTDIR, filename), result)
        else:
            print "No grid found"
