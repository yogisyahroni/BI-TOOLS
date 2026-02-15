import requests
import json

BASE_URL = "http://localhost:8080"
EMAIL = "demo@spectra.id"
PASSWORD = "password123"

def main():
    # 1. Login
    print(f"Logging in as {EMAIL}...")
    try:
        resp = requests.post(f"{BASE_URL}/api/auth/login", json={"email": EMAIL, "password": PASSWORD})
        resp.raise_for_status()
        data = resp.json()
        token = data.get("token")
        if not token:
            print("❌ Login failed: No token in response")
            print(data)
            return
        print("✅ Login successful")
    except Exception as e:
        print(f"❌ Login error: {e}")
        return

    # 2. Generate Embed Token
    print("Generating Embed Token...")
    try:
        headers = {"Authorization": f"Bearer {token}"}
        payload = {
            "dashboard_id": "dash_252f5342-83b3-4f92-9cad-0e1f2031a2b",  # Example ID
            "expiration_minutes": 60,
            "allowed_filters": {},
            "theme": "light",
            "hidden_widgets": []
        }
        resp = requests.post(f"{BASE_URL}/api/embed/token", json=payload, headers=headers)
        resp.raise_for_status()
        embed_data = resp.json()
        embed_token = embed_data.get("token")
        
        if not embed_token:
             print("❌ Embed Token generation failed: No token in response")
             print(embed_data)
             return

        print("\n✅ EMBED TOKEN GENERATED SUCCESSFULLY:")
        print("-" * 50)
        print(embed_token)
        print("-" * 50)
        
    except Exception as e:
        print(f"❌ Embed Token error: {e}")
        if resp:
            print(resp.text)

if __name__ == "__main__":
    main()
