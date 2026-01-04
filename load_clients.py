import asyncio
import websockets
import random
import time

from load_image import image_to_pixel_list

uri = "ws://localhost:8080/ws"

async def drain_messages(websocket):
    try:
        async for _ in websocket:
            pass  # Ignore incoming messages
    except asyncio.CancelledError:
        pass
    except Exception:
        pass

async def client(img, xs, ys):
    try:
        async with websockets.connect(uri, ping_timeout=60, close_timeout=10) as websocket:
            consumer_task = asyncio.create_task(drain_messages(websocket))

            try:
                while True:
                    x, y, color = random.choice(img)
                    x = (x + xs) % 1000
                    y = (y + ys) % 1000

                    data = x << 16 | y << 4 | color
                    msg = data.to_bytes(4, byteorder='big')

                    await websocket.send(msg)
                    await asyncio.sleep(random.uniform(0.5, 1.0))
            finally:
                consumer_task.cancel()

    except Exception as e:
        print(f"Client error: {e}")

async def s_client():
    try:
        async with websockets.connect(uri, ping_timeout=60, close_timeout=10) as websocket:
            consumer_task = asyncio.create_task(drain_messages(websocket))

            try:
                while True:
                    x = random.randint(0, 999)
                    y = random.randint(0, 999)
                    color = random.randint(0, 15)

                    data = x << 16 | y << 4 | color
                    msg = data.to_bytes(4, byteorder='big')

                    await websocket.send(msg)
                    await asyncio.sleep(.5)
            finally:
                consumer_task.cancel()

    except Exception as e:
        print(f"Client error: {e}")

async def main():
    number_of_clients = 5000
    ramp_up_interval = 0.5

	# All credits for their ownership of these images go to their respective creators
	# Images sourced from https://pixilart.com/ and https://tenor.com/
    # I don't gain any benefit from these images, nor do I claim ownership of them
    img_urls = [
        "https://art.pixilart.com/sr5z6be59ee9f2aws3.png",
        "https://art.pixilart.com/sr5zd100fa1044aws3.png",
        "https://art.pixilart.com/sr5z463605d847aws3.png",
        "https://art.pixilart.com/sr5z46709dad2daws3.png",
        "https://media.tenor.com/t3dLLNaI50oAAAAe/cat-cats.png",
        "https://media.tenor.com/XXDVU_	GxHZYAAAAe/joker-speed-trying-not-to-laugh.png",
        "https://art.pixilart.com/sr5z8db621516eaws3.png",
        "https://art.pixilart.com/sr5z6be59ee9f2aws3.png",
    ]
    batch_size = number_of_clients // len(img_urls) # each image gets equal clients

    imgs = []
    for url in img_urls:
        try:
            img_pixels = image_to_pixel_list(url, width=200 ,URL=True)
            imgs.append(img_pixels)
        except Exception:
            pass

    if not imgs:
        print("No images loaded. Exiting.")
        return

    print("Images loaded")
    print(f"Starting {number_of_clients} clients in batches of {batch_size}...")

    xs = random.randint(0, 999)
    ys = random.randint(0, 999)

    running_tasks = []

    for batch_start in range(0, number_of_clients, batch_size):
        batch_end = min(batch_start + batch_size, number_of_clients)
        print(f"Connecting clients {batch_start} to {batch_end}...")

        img = imgs[(batch_start // batch_size) % len(img_urls)]
        xs = (xs + random.randint(-50, 300)) % 1000
        ys = (ys + random.randint(-150, 150)) % 1000

        for i in range(batch_start, batch_end):
            task = asyncio.create_task(client(img, xs, ys))
            running_tasks.append(task)

        await asyncio.sleep(ramp_up_interval)

    # Start Sabotagers
    print("Starting sabotagers...")
    for _ in range(100):
        task = asyncio.create_task(s_client())
        running_tasks.append(task)

    print("All clients connected. Monitoring...")

    # Wait for all tasks to complete (they likely won't unless errored/cancelled)
    await asyncio.gather(*running_tasks, return_exceptions=True)

    print("All clients finished")

if __name__ == "__main__":
	start_time = time.time()
	asyncio.run(main())
	print(f"Total time: {time.time() - start_time:.2f}s")
