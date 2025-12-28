import asyncio
import websockets
import struct

async def client():
	uri = "ws://localhost:8080/ws"
	try:
		async with websockets.connect(uri) as websocket:
			for j in range(5):
				for i in range(200):
					x = 500 + i + (j * 200)
					y = 500 + i + (j * 200)
					color = (i % 4)

					data = (x << 16) | (y << 4) | color
					msg = struct.pack(">I", data)
					await websocket.send(msg)
					await asyncio.sleep(0.01)  # Add 10ms delay to avoid flooding
	except websockets.exceptions.ConnectionClosedError as e:
		print(f"Connection closed: {e}")
	except Exception as e:
		print(f"Error: {e}")

asyncio.run(client())
