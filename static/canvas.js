const colors = [
	[255, 255, 255], // White
	[228, 228, 228], // Light Gray
	[136, 136, 136], // Gray
	[34, 34, 34],    // Black
	[255, 167, 209], // Pink
	[229, 0, 0],     // Red
	[229, 149, 0],   // Orange
	[160, 106, 66],  // Brown
	[229, 217, 0],   // Yellow
	[148, 224, 68],  // Light Green
	[2, 190, 1],     // Green
	[0, 211, 221],   // Cyan
	[0, 131, 199],   // Blue
	[0, 0, 234],     // Dark Blue
	[207, 110, 228], // Lavender
	[130, 0, 128]    // Purple
]

const canvas = document.getElementById("rplace")

async function init() {
	const res = await fetch("/canvas")
	const buffer = await res.arrayBuffer()
	const bytes = new Uint8Array(buffer)

	/** @type {HTMLCanvasElement} */
	const ctx = canvas.getContext("2d")

	/** @type {ImageData} */
	const imgData = ctx.createImageData(1000, 1000)

	for (let i = 0; i < bytes.length; i++) {
		const byte = bytes[i]
		const p1 = (byte >> 4) & 0x0F
		const p2 = byte & 0x0F

		setPixelInImageData(imgData.data, i * 2, p1);
		setPixelInImageData(imgData.data, i * 2 + 1, p2);
	}
	ctx.putImageData(imgData, 0, 0);
}

function setPixelInImageData(data, index, colorIdx) {
	const rgb = colors[colorIdx];
	const i = index * 4;

	data[i] = rgb[0];
	data[i + 1] = rgb[1];
	data[i + 2] = rgb[2];
	data[i + 3] = 255; // Alpha
}

openWebSocket = () => {
	const socket = new WebSocket("ws://localhost:8080/ws");
	socket.binaryType = "arraybuffer";

	socket.onopen = () => {
		console.log("WebSocket connected");
	};

	socket.onerror = (error) => {
		console.error("WebSocket error:", error);
	};

	socket.onclose = () => {
		console.log("WebSocket closed");
	};

	socket.onmessage = (event) => {
		const buffer = new Uint8Array(event.data);
		const ctx = canvas.getContext("2d");

		// check for the special packet
		// This works cuz we append the payload head with out 5 packet
		// so its always 5 or more
		if (buffer.length >= 5 && buffer[0] === 255) {
			const count = (buffer[1] << 24) | (buffer[2] << 16) | (buffer[3] << 8) | buffer[4];
			document.getElementById("clientCount").innerText = count;

			if (buffer.length > 5) {
				for (let i = 5; i < buffer.length; i += 4) {
					const x = (buffer[i] << 8) | buffer[i + 1];
					const yAndColor = (buffer[i + 2] << 8) | buffer[i + 3];
					const y = yAndColor >> 4;
					const colorIdx = yAndColor & 0x0F;

					if (x >= 0 && x < 1000 && y >= 0 && y < 1000) {
						const rgb = colors[colorIdx];
						ctx.fillStyle = `rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})`;
						ctx.fillRect(x, y, 1, 1);
					}
				}
			}
		}
	}
}

init().then(() => {
	openWebSocket();
})
