import numpy as np
import cv2
import sys


def rgb_to_hsv(im):
    return cv2.cvtColor(im, cv2.COLOR_BGR2HSV)


def angle_cos(p0, p1, p2):
    d1, d2 = (p0 - p1).astype('float'), (p2 - p1).astype('float')
    return abs(np.dot(d1, d2) / np.sqrt(np.dot(d1, d1) * np.dot(d2, d2)))

def filter_mask(mask,height, width):
    mask2 = np.zeros((height, width, 1), np.uint8)
    thresh = cv2.threshold(mask, 50, 255, cv2.THRESH_TOZERO)[1]  # 50 too high, 25 too low
    contours, hierarchy = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_NONE)

    l = len(contours)
    print(l)
    if l > 3:
        l = 3

    contours = sorted(contours, key=lambda x: cv2.contourArea(x), reverse=True)
    for i in range(0, l, 1):
         print(l)
         cv2.drawContours(mask2, [contours[i]], -1, (255, 255, 255), -1)

    return mask2

def create_letter_mask(im):
    image_saturation0 = rgb_to_hsv(im)[:, :, 0]  # output
    image_saturation1 = rgb_to_hsv(im)[:, :, 1]  # output

    height, width = image_saturation1.shape
    size = height * width
    mask = np.zeros((height, width, 1), np.uint8)
    kernel = cv2.getStructuringElement(cv2.MORPH_CROSS, (3, 3))

    for layer in [image_saturation0, image_saturation1]:
        for thrs in range(0, 255, 2):
            thresh_white = cv2.threshold(layer, thrs, 255, cv2.THRESH_TOZERO_INV)[1]  # 50 too high, 25 too low
            contours, hierarchy = cv2.findContours(cv2.dilate(thresh_white, kernel, iterations=1),
                                                          cv2.RETR_EXTERNAL,
                                                          cv2.CHAIN_APPROX_NONE)
            for contour in contours:
                # [x, y, w, h] = cv2.boundingRect(contour)
                rect = cv2.minAreaRect(contour)  # basically you can feed this rect into your classifier
                (x, y), (w, h), a = rect  # a - angle
                box = cv2.boxPoints(rect)
                box = np.int0(box)  # turn into ints

                s = w * h
                if s > 0.9 * size:
                    cnt_len = cv2.arcLength(contour, True)
                    cnt = cv2.approxPolyDP(contour, 0.02 * cnt_len, True)
                    if len(cnt) == 4 and cv2.contourArea(cnt) > 1000 and cv2.isContourConvex(cnt):
                        cnt = cnt.reshape(-1, 2)
                        max_cos = np.max([angle_cos(cnt[i], cnt[(i + 1) % 4], cnt[(i + 2) % 4]) for i in range(4)])
                        if max_cos < 0.1:
                            cv2.drawContours(mask, [cnt], -1, (0, 255, 0), 3)

                if h > 1.1 * w:
                    continue

                if h > height / 8 or h < height / 50:
                    continue

                if 0.5 < abs(a):
                    continue

                cv2.drawContours(mask, [box], 0, (255, 255, 255), -1)

    mask = cv2.morphologyEx(mask, cv2.MORPH_CLOSE, np.ones((20,20),np.uint8))
    mask = cv2.morphologyEx(mask, cv2.MORPH_OPEN, np.ones((20,20),np.uint8))

    mask = filter_mask(mask,height, width)

    return mask


def equlColor(img):
    ycrcb = cv2.cvtColor(img, cv2.COLOR_BGR2YCR_CB)
    channels = cv2.split(ycrcb)
    cv2.equalizeHist(channels[0], channels[0])
    cv2.merge(channels, ycrcb)
    cv2.cvtColor(ycrcb, cv2.COLOR_YCR_CB2BGR, img)
    return img


def main():
    name = sys.argv[1]

    im = cv2.imread(name)
    blurred = cv2.GaussianBlur(im, (5, 5), 0)
    mask = create_letter_mask(blurred)

    dst = cv2.inpaint(im, mask, 3, cv2.INPAINT_TELEA)
    res = equlColor(dst)
    #im1 = cv2.imread(name)
    #cv2.imwrite(name, mask)
    im = cv2.imread(name)
    res = np.hstack((im, res))

    cv2.imwrite(name, res)


if __name__ == '__main__':
    main()
