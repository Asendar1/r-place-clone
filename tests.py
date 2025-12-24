import asyncio
import websockets


async def listen_client():
	uri = "ws://localhost:8080/ws"
	async with websockets.connect(uri) as websocket:
		while True:
			try:
				response = await websocket.recv()
				print(f"Received from server: {response}")
			except websockets.ConnectionClosed:
				print("Connection closed")
				break

async def client():
	uri = "ws://localhost:8080/ws"
	async with websockets.connect(uri) as websocket:
		await websocket.send("hello")

asyncio.run(listen_client())
asyncio.run(client())

