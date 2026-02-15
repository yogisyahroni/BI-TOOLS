import requests
import json

url = "http://localhost:8080/api/auth/register"
payload = {
    "email": "test_debug_reg_py@example.com",
    "password": "TestPass123!",
    "username": "test_debug_reg_py",
    "fullName": "Test Debug Py"
}
headers = {
    "Content-Type": "application/json"
}

try:
    response = requests.post(url, json=payload, headers=headers)
    print("Status Code:", response.status_code)
    print("Response Body:", response.text)
    
    with open("error.log", "w") as f:
        f.write(response.text)
except Exception as e:
    print(f"Error: {e}")
