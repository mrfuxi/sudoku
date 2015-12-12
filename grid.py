#!/usr/bin/env python
import glob
from collections import defaultdict
from os import path, mkdir
from shutil import rmtree
import numpy as np

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
            f = abs(abs(point-points[-1]) - step)/step
            if f >= 0.2:
                break
            fit += f/9.0

        points.append(point)

    return (1-fit), points


def prepare_point_distances(points):
    max_p = points[-1]
    ld = int(max_p+1)
    distances = [(max_p, max_p)] * ld
    for point in points:
        idx = int(point)
        distances[idx] = (0, point)
        for i in xrange(1, idx+1):
            if (i, point) < distances[idx - i]:
                distances[idx - i] = (i, point)
            else:
                break

        for i in xrange(1, ld - idx):
            if (i, point) < distances[idx + i]:
                distances[idx + i] = (i, point)
            else:
                break

    return [d[1] for d in distances]


def linear_distances(lines, divider_line):
    if len(lines) < 10:
        yield 0.0, []

    intersections = []
    for line in lines:
        _, point = intersection(line, divider_line)
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
    vertical.sort()
    horizontal.sort()

    lines_v = defaultdict(list)
    for fit in (linear_distances(vertical, h) for h in horizontal):
        for score, line in fit:
            lines_v[tuple(line)].append(score)

    lines_h = defaultdict(list)
    for fit in (linear_distances(horizontal, v) for v in vertical):
        for score, line in fit:
            lines_h[tuple(line)].append(score)

    lines_v = [(l, np.mean(s)) for l, s in lines_v.items()]
    lines_h = [(l, np.mean(s)) for l, s in lines_h.items()]

    grids = []
    for h, hs in lines_h[:3]:
        for v, vs in lines_v[:3]:
            grids.append((hs*vs, (h, v)))

    return grids


def evaluate_grids(img, grids):
    best = (10**10, None)
    for i, (line_score, grid) in enumerate(grids):
        lines_h, lines_v = grid
        fragments = []
        for h in lines_h:
            _, point_a = intersection(h, lines_v[0])
            _, point_b = intersection(h, lines_v[-1])
            fragments.append((point_a, point_b))

        for v in lines_v:
            _, point_a = intersection(v, lines_h[0])
            _, point_b = intersection(v, lines_h[-1])
            fragments.append((point_a, point_b))

        fragment_image = visualize.draw_fragments(
            img, fragments, color=255, width=2
        )
        masked_image = cv2.bitwise_and(img, fragment_image)
        score = masked_image.sum()

        # max dark ink, so minimize it
        best = min(best, (score*line_score, grid))

    return best


def find_grid(image, fn):
    img = process.pre_process(image)
    img_grey = process.gray_image(image)

    lines = process.find_lines(img, 100)
    dedup = remove_duplicate_lines(lines, img.shape)

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

        vertical, horizontal = lines_with_similar_angle(line_class, angle)

        if len(vertical) < 10 or len(horizontal) < 10:
            continue

        grids = possible_grids(horizontal, vertical)
        score, grid = evaluate_grids(img_grey, grids)
        best = max(best, (score, grid))

    return best


def grid_corners(grid, row=None, col=None):
    if not row:
        row = (0, -1)

    if not col:
        col = (0, -1)

    horizontal, vertical = grid

    h1 = horizontal[row[0]]
    h2 = horizontal[row[1]]
    v1 = vertical[col[0]]
    v2 = vertical[col[1]]

    return (
        intersection(h1, v1)[1],
        intersection(h1, v2)[1],
        intersection(h2, v2)[1],
        intersection(h2, v1)[1],
    )


def cut_square_from_image(img, corners, size):
    quad_pts = np.array([(0, 0), (size, 0), (size, size), (0, size)], dtype=np.float32)
    transmtx = cv2.getPerspectiveTransform(corners, quad_pts)
    return cv2.warpPerspective(img, transmtx, (size, size))


def cut_grid(img, grid, size=360):
    corners = np.array(grid_corners(grid))
    return cut_square_from_image(img, corners, size)


def cut_cells_from_grid(img, grid, size=32):
    grey_img = process.gray_image(img)
    cells = []
    for row in range(0, 9):
        for col in range(0, 9):
            corners = np.array(grid_corners(grid, row=(row, row+1), col=(col, col+1)))
            cell = cut_square_from_image(grey_img, corners, size)
            _, cell = cv2.threshold(cell, 0, 255, cv2.THRESH_BINARY+cv2.THRESH_OTSU)
            cells.append(cell)
    return cells


def cell_to_feature_vector(cell, size=4):
    features = []
    cell_size = cell.shape[0]

    for i in range(0, cell_size, size):
        for j in range(0, cell_size, size):
            region = cell[i:i+size, j:j+size]
            features.append(size*size - np.count_nonzero(region))

    return features

if __name__ == '__main__':
    rmtree(OUTDIR, ignore_errors=True)
    mkdir(OUTDIR)

    from time import time

    for filename in sorted(glob.glob('examples/*.png')):
        filename = path.basename(filename)
        img = process.get_example_image(filename)
        t0 = time()
        score, grid = find_grid(img, filename)

        if grid:
            cut = cut_grid(img, grid)
            cv2.imwrite("{}/cut_{}".format(OUTDIR, filename), cut)
            result = visualize.draw_lines(img, grid[0] + grid[1], thickness=2)
            cv2.imwrite("{}/{}".format(OUTDIR, filename), result)
            for i, cell in enumerate(cut_cells_from_grid(img, grid)):
                cv2.imwrite("{}/cell_{}_{}".format(OUTDIR, i, filename), cell)
            cell_to_feature_vector(cell)
        else:
            print "No grid found", filename

        t1 = time()
        print t1-t0
