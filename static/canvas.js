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

//#region Initialization
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
//#endregion

function setPixelInImageData(data, index, colorIdx) {
	const rgb = colors[colorIdx];
	const i = index * 4;

	data[i] = rgb[0];
	data[i + 1] = rgb[1];
	data[i + 2] = rgb[2];
	data[i + 3] = 255; // Alpha
}

const socket = new WebSocket("ws://localhost:8080/ws");

openWebSocket = () => {
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

const wrapper = document.getElementById("canvasWrapper");


let scale = 1;
const minScale = 1;
const maxScale = 10;
const baseWidth = 1000;
const baseHeight = 1000;

canvas.addEventListener("wheel", (e) => {
    e.preventDefault();

    const rect = wrapper.getBoundingClientRect();

    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Thank you A.I god forbid i learn JS
    const contentX = (wrapper.scrollLeft + mouseX) / scale;
    const contentY = (wrapper.scrollTop + mouseY) / scale;

    const delta = e.deltaY < 0 ? 1.1 : 0.9;
    const newScale = Math.min(maxScale, Math.max(minScale, scale * delta));

    if (newScale !== scale) {
        scale = newScale;

        // Resize BOTH the container and the canvas
        container.style.width = `${baseWidth * scale}px`;
        container.style.height = `${baseHeight * scale}px`;
        canvas.style.width = `${baseWidth * scale}px`;
        canvas.style.height = `${baseHeight * scale}px`;

        wrapper.scrollLeft = (contentX * scale) - mouseX;
        wrapper.scrollTop = (contentY * scale) - mouseY;
    }
}, { passive: false });

let isPanning = false;
let startX, startY;
let startScrollLeft, startScrollTop;


wrapper.addEventListener("mousedown", (e) => {
    isPanning = true;
    wrapper.style.cursor = "grabbing";

    startX = e.pageX;
    startY = e.pageY;

    startScrollLeft = wrapper.scrollLeft;
    startScrollTop = wrapper.scrollTop;
});

window.addEventListener("mousemove", (e) => {
    if (!isPanning) return;

    const walkX = e.pageX - startX;
    const walkY = e.pageY - startY;

    wrapper.scrollLeft = startScrollLeft - walkX;
    wrapper.scrollTop = startScrollTop - walkY;
});

window.addEventListener("mouseup", () => {
    isPanning = false;
    wrapper.style.cursor = "default";
});

let selectedColor = 0;

const container = document.getElementById("canvasContainer"); // New reference
const highlight = document.createElement("div");

highlight.style.position = "absolute";
highlight.style.border = "1px solid white";
highlight.style.pointerEvents = "none";
highlight.style.boxSizing = "border-box";
highlight.style.zIndex = "10";
highlight.style.display = "none"; // Start hidden
container.appendChild(highlight); // Attach to container!

wrapper.addEventListener("mousemove", (e) => {
    const rect = wrapper.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Calculate the pixel coordinate (0-999)
    const x = Math.floor((wrapper.scrollLeft + mouseX) / scale);
    const y = Math.floor((wrapper.scrollTop + mouseY) / scale);

    if (x >= 0 && x < 1000 && y >= 0 && y < 1000) {
        highlight.style.display = "block";
        highlight.style.width = `${scale}px`;  // Match pixel size
        highlight.style.height = `${scale}px`;
        highlight.style.left = `${x * scale}px`; // Exact pixel position
        highlight.style.top = `${y * scale}px`;
    } else {
        highlight.style.display = "none";
    }
});

wrapper.addEventListener("click", (e) => {
    // If we were just panning don't paint
    if (Math.abs(e.pageX - startX) > 5 || Math.abs(e.pageY - startY) > 5) return;

    const rect = wrapper.getBoundingClientRect();
    const x = Math.floor((wrapper.scrollLeft + (e.clientX - rect.left)) / scale);
    const y = Math.floor((wrapper.scrollTop + (e.clientY - rect.top)) / scale);
	console.log(`Clicked at canvas coords: (${x}, ${y}) with color ${selectedColor}`);
    if (x >= 0 && x < 1000 && y >= 0 && y < 1000) {
        paintPixel(x, y, selectedColor);
    }
});

function paintPixel(x, y, colorIdx) {
    const buffer = new ArrayBuffer(4);
    const view = new DataView(buffer);
    view.setUint16(0, x);
    view.setUint16(2, (y << 4) | (colorIdx & 0x0F));

    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(buffer);
    }
}

const buttons = document.querySelectorAll("#pallete button");
buttons.forEach((button, index) => {
	button.addEventListener("click", () => {
		selectedColor = index;
		buttons.forEach(btn => btn.style.outline = "none");
		button.style.outline = "3px solid white";
	});
});
