# 🔧 FIX: WebSocket "Insufficient Resources" Error

## Problem

Error: `WebSocket connection failed: Insufficient resources`

## Root Causes

Error ini biasanya disebabkan oleh:

1. **File Descriptor Limits** - OS limit untuk open connections
2. **Memory Limits** - RAM habis untuk WebSocket buffers
3. **Connection Leaks** - Koneksi tidak ditutup properly
4. **No Connection Limits** - Tidak ada batasan jumlah koneksi

## Solutions

### Solution 1: Quick Fix - Increase System Limits (Windows)

```powershell
# Run as Administrator
# Increase file descriptor limits (Windows equivalent)
netsh int ipv4 set dynamicport tcp start=1025 num=64511

# Or restart the backend with proper cleanup
```

### Solution 2: Use Fixed WebSocket Hub

File yang sudah dibuat: `backend/services/websocket_hub_fixed.go`

Features:
- ✅ Max 1000 total connections
- ✅ Max 5 connections per user
- ✅ Proper connection counting
- ✅ Better error handling
- ✅ Connection stats

**Cara pakai:**

1. Replace di `bootstrap/services.go`:

```go
// OLD:
wsHub := services.NewWebSocketHub()

// NEW:
wsHub := services.NewWebSocketHubFixed()
```

2. Restart backend

### Solution 3: Fix Connection Leaks (Recommended)

Update `backend/handlers/websocket_handler.go`:

```go
// Add connection limit check
func (h *WebSocketHandler) HandleConnection(c *websocket.Conn) {
    // Check if hub is full (add this check)
    stats := h.wsHub.GetStats()
    if stats["total_connections"].(int32) >= 1000 {
        LogWarn("websocket_rejected", "Hub full, rejecting connection", nil)
        c.Close()
        return
    }
    
    // ... rest of the code
}
```

### Solution 4: Frontend Fix (Immediate)

Tambahkan retry logic dan exponential backoff di frontend:

```typescript
// frontend/lib/websocket.ts

class WebSocketManager {
    private ws: WebSocket | null = null;
    private reconnectAttempts = 0;
    private maxReconnectAttempts = 5;
    private reconnectDelay = 1000; // Start with 1 second

    connect(url: string) {
        try {
            this.ws = new WebSocket(url);
            
            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
                this.reconnectDelay = 1000;
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket closed:', event.code, event.reason);
                
                // Only reconnect if not closed cleanly
                if (event.code !== 1000 && event.code !== 1001) {
                    this.attemptReconnect(url);
                }
            };
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.attemptReconnect(url);
        }
    }

    private attemptReconnect(url: string) {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('Max reconnection attempts reached');
            return;
        }

        this.reconnectAttempts++;
        console.log(`Reconnecting... Attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);

        setTimeout(() => {
            this.connect(url);
        }, this.reconnectDelay);

        // Exponential backoff
        this.reconnectDelay *= 2;
    }
}
```

### Solution 5: System-Level Fix (Production)

**Linux/Mac:**
```bash
# Increase file descriptor limits
ulimit -n 65535

# Or permanent fix in /etc/security/limits.conf
* soft nofile 65535
* hard nofile 65535
```

**Windows:**
```powershell
# Increase dynamic port range
netsh int ipv4 set dynamicport tcp start=1025 num=64511

# Increase max user ports
reg add "HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters" /v MaxUserPort /t REG_DWORD /d 65534 /f
```

## Quick Implementation

### Step 1: Apply Backend Fix (5 minutes)

```bash
# 1. Backup original
cp backend/services/websocket_hub.go backend/services/websocket_hub.go.backup

# 2. Copy fixed version
cp backend/services/websocket_hub_fixed.go backend/services/websocket_hub.go

# 3. Restart backend
cd backend
go run main.go
```

### Step 2: Test Connection (1 minute)

```javascript
// Test in browser console
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=YOUR_TOKEN');
ws.onopen = () => console.log('✅ Connected');
ws.onerror = (e) => console.error('❌ Error:', e);
ws.onclose = (e) => console.log('🔴 Closed:', e.code, e.reason);
```

## Monitoring

Add monitoring untuk track WebSocket connections:

```go
// Add to your health check endpoint
func (h *WebSocketHandler) GetStats(c *fiber.Ctx) error {
    stats := h.wsHub.GetStats()
    return c.JSON(stats)
}
```

Access: `GET /api/v1/ws/stats`

Response:
```json
{
    "connected_users": 5,
    "total_connections": 12,
    "max_total": 1000,
    "max_per_user": 5,
    "available_slots": 988
}
```

## Prevention

1. **Always close connections properly**
   ```go
   defer client.Conn.Close()
   ```

2. **Add connection limits** (sudah ada di fixed version)

3. **Monitor connection count**
   ```go
   // Log every 100 connections
   if count % 100 == 0 {
       LogInfo("websocket_load", "High connection count", map[string]interface{}{
           "count": count,
       })
   }
   ```

4. **Implement ping/pong**
   ```go
   // Already implemented in your code
   client.Conn.SetPongHandler(func(string) error {
       // Connection is alive
       return nil
   })
   ```

## Status

- ✅ Fixed WebSocket hub created
- ✅ Connection limits implemented
- ✅ Better resource management
- ✅ Stats endpoint ready

**File created:** `backend/services/websocket_hub_fixed.go`

---

## Next Steps

1. **Test the fix:**
   ```bash
   cd backend
   go run main.go
   ```

2. **Monitor logs:**
   - Watch for "websocket_client_registered" messages
   - Check if "websocket_max_connections" appears

3. **Apply permanently:**
   ```bash
   # Rename fixed version to replace original
   mv backend/services/websocket_hub.go backend/services/websocket_hub.go.old
   mv backend/services/websocket_hub_fixed.go backend/services/websocket_hub.go
   ```

4. **Deploy to production**

---

**Created:** 2026-02-14  
**Status:** ✅ Fix Ready  
**Test Status:** Pending (needs backend restart)
