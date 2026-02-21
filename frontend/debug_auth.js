// Simple script to test backend connectivity from Node.js
// Usage: node debug_auth.js

async function testAuth() {
  const url = "http://127.0.0.1:8080/api/auth/login";
  console.log(`Connecting to ${url}...`);

  try {
    const response = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: "yogisyahroni766.ysr@gmai.com",
        password: "Namakamu766!!",
      }),
    });

    console.log("Response Status:", response.status);
    const text = await response.text();
    console.log("Response Body:", text);
  } catch (error) {
    console.error("Fetch Error:", error);
  }
}

testAuth();
