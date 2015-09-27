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


def draw_lines(img, lines):
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
