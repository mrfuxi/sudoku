import cv2
import numpy as np


def draw_points(img, points):
    """
    Draws lines intersections.
    To visualize state of processing
    """

    cpy = img.copy()
    color = (255, 0, 0)

    for point in points:
        cv2.circle(cpy, tuple(point), 2, color, thickness=-1)

    return cpy


def draw_lines(img, lines, color=None, thickness=2, rgb=True, draw_on_empty=False):
    """
    Draws lines on image.
    To visualize state of processing
    """
    if draw_on_empty:
        cpy = np.empty_like(img)
        cpy.fill(0)
    else:
        cpy = img.copy()

    if rgb:
        if not color:
            color = (0, 255, 0)

        if len(cpy.shape) == 2:
            cpy = cv2.cvtColor(cpy, cv2.COLOR_GRAY2RGB)

    elif not color:
        color = 255

    if cpy.max() <= 1:
        cpy *= 255

    img_rect = (0, 0, img.shape[1], img.shape[0])

    for rho, theta in lines:
        a = np.cos(theta)
        b = np.sin(theta)
        x0 = a*rho
        y0 = b*rho
        x1 = int(x0 + 10000*(-b))
        y1 = int(y0 + 10000*(a))
        x2 = int(x0 - 10000*(-b))
        y2 = int(y0 - 10000*(a))

        in_view, start, end = cv2.clipLine(img_rect, (x1, y1), (x2, y2))
        if not in_view:
            continue

        cv2.line(cpy, start, end, color, thickness)

    return cpy


def draw_fragment_values(img, fragments):
    cpy = np.empty_like(img)
    cpy.fill(0)
    m = np.max(fragments.values())
    for (point_a, point_b), score in fragments.items():
        cv2.line(cpy, point_a, point_b, int(255*score/m), 5)

    return cpy
