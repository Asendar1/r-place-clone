import asyncio
import websockets
import random
import time

async def client(client_id):
    uri = "ws://localhost:8080/ws"
    try:
        async with websockets.connect(
            uri,
            ping_interval=None,  # Disable client pings
            ping_timeout=None,   # We'll handle server pings
            close_timeout=10
        ) as websocket:

            # Send 100 random pixels
            for _ in range(100):
                x = random.randint(0, 999)
                y = random.randint(0, 999)
                color = random.randint(0, 15)  # Fixed: was 0-16

                data = (x << 16) | (y << 4) | color
                await websocket.send(data.to_bytes(4, byteorder='big'))

                # Don't spam - respect rate limit
                await asyncio.sleep(random.uniform(0.5, 1.0))

            # Keep connection alive for a bit to receive updates
            await asyncio.sleep(5)

    except websockets.exceptions.ConnectionClosedError as e:
        print(f"Client {client_id} - Connection closed: {e}")
    except asyncio.TimeoutError:
        print(f"Client {client_id} - Timeout during handshake")
    except Exception as e:
        print(f"Client {client_id} - Error: {e}")

async def main():
    number_of_clients = 10000
    batch_size = 100  # Connect in batches

    print(f"Starting {number_of_clients} clients in batches of {batch_size}...")

    for batch_start in range(0, number_of_clients, batch_size):
        batch_end = min(batch_start + batch_size, number_of_clients)
        print(f"Connecting clients {batch_start} to {batch_end}...")

        tasks = [client(i) for i in range(batch_start, batch_end)]
        await asyncio.gather(*tasks, return_exceptions=True)

        # Brief pause between batches
        await asyncio.sleep(0.5)

    print("All clients finished")

if __name__ == "__main__":
    start_time = time.time()
    asyncio.run(main())
    print(f"Total time: {time.time() - start_time:.2f}s")
