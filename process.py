import cv2
import numpy as np


def get_example_image(name):
    return cv2.imread("examples/{}".format(name))


def gray_image(img):
    return cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)


def binarize(img):
    """
    Initial threshold to get binary image
    """

    window = img.shape[0]/10
    if window % 2 == 0:
        window += 1

    th = cv2.adaptiveThreshold(
        img,
        maxValue=1,
        adaptiveMethod=cv2.ADAPTIVE_THRESH_MEAN_C,
        thresholdType=cv2.THRESH_BINARY_INV,
        blockSize=window,
        C=0,
    )

    return th


def remove_blobs_body(img):
    """
    Removes body of regions over 1/20 of image width
    """

    window = (img.shape[0]/20)
    if window % 2 == 0:
        window += 1

    th = cv2.adaptiveThreshold(
        img,
        maxValue=1,
        adaptiveMethod=cv2.ADAPTIVE_THRESH_MEAN_C,
        thresholdType=cv2.THRESH_BINARY,
        blockSize=window,
        C=0,
    )

    return th


def pre_process(img):
    """
    Prepared original image for actual work
    """

    grey = gray_image(img)
    binary = binarize(grey)
    deblobbed = remove_blobs_body(binary)

    return deblobbed


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


def find_lines(img, limit=100):
    threshold = 0
    rho = 1
    theta = np.pi/180

    lines = cv2.HoughLines(img, rho=rho, theta=theta, threshold=threshold)
    return lines[:limit]


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
            points[key] = tuple(point.flatten())

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

            ok, point = intersection(line_a, line_b)
            if not ok:
                continue

            in_view = point_in_view(point, img_shape)
            similar = similar_angle(line_a, line_b, min_ang_diff)

            if in_view and similar:
                to_remove.add(max(i, j))

    cleaned = [line for i, line in enumerate(lines) if i not in to_remove]
    return cleaned


def draw_points(img, points):
    cpy = img.copy()
    color = (255, 0, 0)

    for point in points:
        cv2.circle(cpy, point, 2, color, thickness=-1)

    return cpy


def draw_hlines(img, lines):
    """
    Draws lines on image.
    To visualize state of processing
    """

    cpy = img.copy()

    for line in lines:
        rho, theta = line[0]
        a = np.cos(theta)
        b = np.sin(theta)
        x0 = a*rho
        y0 = b*rho
        x1 = int(x0 + 1000*(-b))
        y1 = int(y0 + 1000*(a))
        x2 = int(x0 - 1000*(-b))
        y2 = int(y0 - 1000*(a))

        cv2.line(cpy, (x1, y1), (x2, y2), (0, 0, 255), 2)

    return cpy
