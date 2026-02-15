
const fs = require('fs');

async function getToken() {
    const url = 'http://localhost:8080/api/auth/login';
    // Try separate credentials if the default one fails
    const credsList = [
        { email: 'test@example.com', password: 'Password123!' },
        { email: 'admin@example.com', password: 'password' },
        { email: 'user@example.com', password: 'password' }
    ];

    for (const creds of credsList) {
        console.log(`Trying login with ${creds.email}...`);
        try {
            const response = await fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(creds)
            });

            if (response.ok) {
                const data = await response.json();
                if (data.token) {
                    console.log('TOKEN_SUCCESS');
                    fs.writeFileSync('token.txt', data.token);
                    return;
                }
            } else {
                console.log(`Failed: ${response.status}`);
            }
        } catch (e) {
            console.log('Error:', e.message);
        }
    }
    console.log('TOKEN_FAILED');
}

getToken();
