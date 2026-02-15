
async function login() {
    const url = 'http://localhost:8080/api/auth/login';
    const credentials = {
        email: 'test@example.com',
        password: 'Password123!'
    };

    console.log(`Logging in user: ${credentials.email}`);

    try {
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(credentials),
        });

        const status = response.status;
        console.log(`Response Status: ${status}`);

        const data = await response.json();
        console.log('Response Body:', JSON.stringify(data, null, 2));

        if (response.ok) {
            console.log('✅ Login successful!');
            if (data.token) {
                console.log('Token received:', data.token.substring(0, 20) + '...');
            }
        } else {
            console.error('❌ Login failed.');
        }
    } catch (error) {
        console.error('❌ Error logging in:', error);
    }
}

login();
