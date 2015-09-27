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


def find_lines(img, limit=100):
    threshold = 0
    rho = 1
    theta = np.pi/180

    lines = cv2.HoughLines(img, rho=rho, theta=theta, threshold=threshold)
    return lines[:limit]
