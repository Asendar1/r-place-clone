# High-Concurrency Pixel Engine (r/place Clone)

![Cat-Gif-ezgif com-optimize](https://github.com/user-attachments/assets/64ac5b4d-e818-4359-aa64-260015be92ea)

*5,000 coordinated bots drawing the staring cat vs. 100 saboteurs — all in real-time*

🚀 **Performance at a Glance**

- **Stress Test Load**: 5,000 artist bots + 100 saboteurs (5,100+ total clients)  
- **Canvas Size**: 1,000,000 pixels (1000×1000)  
- **Backend Resource Limits**: Docker container capped at **2 cores / 1GB RAM**

## 🛠️ Technical Architecture

### 1. Sharded WebSocket Hub (Go)
To eliminate global lock contention in high-frequency real-time systems, I designed a **12-shard hub** architecture.

- Each shard manages its own client set with RWMutex  
- Registration, unregistration, and broadcasting scaled linearly  
- Goroutines, channels, and atomic operations for lock-free client counting  
- Batched updates every 100ms to prevent network saturation

**Result**: Near-zero contention even with thousands of simultaneous pixel placements.

### 2. Optimized Persistence (Redis)
Used **Redis BITFIELD (u4)** for pixel storage instead of traditional strings or hashes.

- **O(1)** SET and GET operations per pixel  
- Entire canvas stored in **~500KB**  
- Autosave every 5 minutes with timestamped backups and recovery on restart

### 3. Binary Protocol Over WebSockets
Custom 4-byte message format:

- X: 12 bits (0–999)  
- Y: 12 bits (0–999)  
- Color: 4 bits (0–15)

Eliminates JSON overhead, reduces bandwidth, and enables ultra-fast frontend parsing.

## 🧪 Stress Testing & The "Bot War"

Python asyncio suite simulating real user behavior:

- **5,000 Artist Bots**: Coordinated placement to draw full 1000×1000 images (including the staring cat)  
- **100 Saboteurs**: Random high-frequency pixel spam to create maximum entropy

**Observation during peak load**:
- Go backend CPU usage: ~22%  
- Total system CPU: ~90–100%  
- Bottleneck: WSL2's Vmmem (network bridge) saturating under 5,000+ sockets

**Conclusion**: The Go engine and architecture remained efficient — the limit was host networking, not application logic.

## 💻 Frontend Interactivity

Vanilla JS canvas with full user control:

- Smooth **zoom** and **pan** via CSS transforms  
- Direct binary WebSocket parsing → immediate pixel updates  
- Live client counter and responsive color picker  
- No frame drops even under thousands of updates/sec

## 🏗️ Building & Running

```bash
# Clone the repo
git clone "https://github.com/Asendar1/r-place-clone"
cd r-place-clone

# Start Redis + Go backend (Docker)
docker-compose up --build

# Open in browser (Try it before the bots do :D)
http://localhost:8080/play

# Setup python venv for the tests  (use python3 instead if using linux)
python -m venv venv

# instal dependices
pip install -r requirements.txt

# Launch the Bot War
python load_clients.py
