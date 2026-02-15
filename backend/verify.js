// const fetch = require('node-fetch'); // Built-in in Node 18+

async function verifyEmail() {
    const token = '263cdcdc-566e-4151-ac08-06a483875cf7';
    const url = `http://localhost:8080/api/auth/verify-email?token=${token}`;

    console.log(`Verifying email with token: ${token}`);
    console.log(`URL: ${url}`);

    try {
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        const status = response.status;
        console.log(`Response Status: ${status}`);

        const data = await response.json();
        console.log('Response Body:', JSON.stringify(data, null, 2));

        if (response.ok) {
            console.log('✅ Email verification successful!');
        } else {
            console.error('❌ Email verification failed.');
        }
    } catch (error) {
        console.error('❌ Error verifying email:', error);
    }
}

verifyEmail();
