import glob
from os import path, mkdir
from shutil import rmtree

import process
from low_structures import (
    remove_very_close_lines,
    remove_disjonted_lines,
    remove_duplicate_lines,
    put_lines_into_buckets,
    generate_angle_buckets,
)
import visualize
import cv2


def find_grid(image):
    # img_org = process.get_example_image(filename)  # 's6.png'
    img = process.pre_process(image)

    lines = process.find_lines(img, 100)
    dedup = remove_duplicate_lines(lines, 15, img.shape)

    bucket_size = 90.0/5.0
    buckets = generate_angle_buckets(
        bucket_size, step=bucket_size/2.0, ortogonal=True
    )
    bucketed_lines = put_lines_into_buckets(buckets, dedup)

    found = False
    fragments = None
    for _, line_class in bucketed_lines:
        not_close = remove_very_close_lines(line_class, img.shape)
        connected, fragments = remove_disjonted_lines(img, not_close)
        # connected = line_class

        if len(line_class) >= 20:
            found = True
            break

    return found, connected, fragments

if __name__ == '__main__':
    out_dir = 'example_out'

    rmtree(out_dir, ignore_errors=True)
    mkdir(out_dir)

    for filename in glob.glob('examples/*.png'):
        filename = path.basename(filename)
        print filename
        img = process.get_example_image(filename)
        found, lines, fragments = find_grid(img)

        if found:
            result = visualize.draw_lines(img, lines, thickness=2)
            cv2.imwrite("{}/{}".format(out_dir, filename), result)

            if fragments:
                result = visualize.draw_fragment_values(img, fragments)
                cv2.imwrite("{}/fragmetns_{}".format(out_dir, filename), result)
