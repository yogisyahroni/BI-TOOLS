import requests
import json
import hmac
import hashlib
import time

BASE_URL = "http://localhost:8080/api"
LOGIN_EMAIL = "demo@spectra.id"
LOGIN_PASSWORD = "password123"

def login():
    url = f"{BASE_URL}/auth/login"
    payload = {
        "email": LOGIN_EMAIL,
        "password": LOGIN_PASSWORD
    }
    response = requests.post(url, json=payload)
    if response.status_code == 200:
        return response.json()["token"]
    else:
        print(f"Login failed: {response.text}")
        return None

def create_webhook(token):
    url = f"{BASE_URL}/webhooks"
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "name": "Test Webhook",
        "url": "https://webhook.site/#!/view/b8b8b8b8-b8b8-b8b8-b8b8-b8b8b8b8b8b8", # Replace with a real webhook.site URL for manual testing if needed, or use a local listener
        "events": ["test.event", "user.signup"],
        "headers": json.dumps({"X-Custom-Header": "InsightEngine"})
    }
    # For automated testing, we might want a local listener, but for now let's just create it
    # We will use "http://localhost:8081/webhook" and run a simple listener if we want to verify delivery automatically
    payload["url"] = "http://localhost:8081/webhook"
    
    response = requests.post(url, json=payload, headers=headers)
    if response.status_code == 201:
        print("Webhook created successfully")
        return response.json()
    else:
        print(f"Failed to create webhook: {response.text}")
        return None

def test_webhook_dispatch(token, webhook_id):
    url = f"{BASE_URL}/webhooks/{webhook_id}/test"
    headers = {"Authorization": f"Bearer {token}"}
    payload = {
        "event": "test.event",
        "payload": {"data": "Hello World"}
    }
    response = requests.post(url, json=payload, headers=headers)
    if response.status_code == 200:
        print("Test event dispatched")
    else:
        print(f"Failed to dispatch test event: {response.text}")

def listener():
    from http.server import BaseHTTPRequestHandler, HTTPServer
    
    class WebhookListener(BaseHTTPRequestHandler):
        def do_POST(self):
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            print("\n--- Webhook Received ---")
            print(f"Headers: {self.headers}")
            print(f"Body: {post_data.decode('utf-8')}")
            
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b"OK")
            
    server_address = ('', 8081)
    httpd = HTTPServer(server_address, WebhookListener)
    print("Starting local webhook listener on port 8081...")
    # Handle one request then exit for this simple test script, or run in thread
    httpd.handle_request()

if __name__ == "__main__":
    # Start listener in background or separate process? 
    # For simplicity, let's run listener in a thread
    import threading
    listener_thread = threading.Thread(target=listener)
    listener_thread.daemon = True
    listener_thread.start()
    
    # Wait for listener to start
    time.sleep(1)
    
    token = login()
    if token:
        webhook = create_webhook(token)
        if webhook:
            print(f"Webhook ID: {webhook['id']}")
            test_webhook_dispatch(token, webhook['id'])
            
            # Wait for delivery
            time.sleep(2)
