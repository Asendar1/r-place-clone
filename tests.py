import asyncio
from asyncio import tasks
import websockets
import struct
import random


async def client(client_id):
    uri = "ws://localhost:8080/ws"
    try:
        async with websockets.connect(uri) as websocket:
            for _ in range(100):
                x = random.randint(0, 999)
                y = random.randint(0, 999)
                color = random.randint(0, 16)

                data = (x << 16) | (y << 4) | color
                await websocket.send(data.to_bytes(4, byteorder='big'))
                await asyncio.sleep(random.uniform(0.1, 0.5))
    except websockets.exceptions.ConnectionClosedError as e:
        print(f"Connection closed: {e}")
    except Exception as e:
        print(f"Error: {e}")


async def main():
    number_of_clients = 10000
    tasks = []
    for i in range(number_of_clients):
        tasks.append(client(i))
    await asyncio.gather(*tasks)

asyncio.run(main())
