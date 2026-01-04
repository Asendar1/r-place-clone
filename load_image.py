from PIL import Image
import requests
from io import BytesIO
import math

PALETTE = [
	(255, 255, 255), (211, 211, 211), (128, 128, 128), (0, 0, 0),
	(255, 192, 203), (255, 0, 0),     (255, 165, 0),   (165, 42, 42),
	(255, 255, 0),   (144, 238, 144), (0, 128, 0),     (0, 255, 255),
	(0, 0, 255),     (0, 0, 139),     (230, 230, 250), (128, 0, 128)
]

def get_closest_color(rgb):
	r, g, b = rgb[:3]
	min_dist = float('inf')
	closest_idx = 0
	for i, color in enumerate(PALETTE):
		dist = math.sqrt((r - color[0])**2 + (g - color[1])**2 + (b - color[2])**2)
		if dist < min_dist:
			min_dist = dist
			closest_idx = i
	return closest_idx

def image_to_pixel_list(image_path, width=100, URL=False):
	if URL:
		response = requests.get(image_path)
		img = Image.open(BytesIO(response.content)).convert('RGB')
	else:
		img = Image.open(image_path).convert('RGB')
	img = img.resize((width, int(width * img.height / img.width)))

	pixels = []
	for y in range(img.height):
		for x in range(img.width):
			color_idx = get_closest_color(img.getpixel((x, y)))
			pixels.append((x, y, color_idx))
	return pixels
