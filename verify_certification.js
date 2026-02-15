const axios = require('axios');
const { execSync } = require('child_process');

const API_URL = 'http://localhost:8080/api';
const TIMESTAMP = Date.now();
const EMAIL = `admin_${TIMESTAMP}@test.com`;
const PASSWORD = 'password123';

async function main() {
    // 1. Register or Login
    let token;
    try {
        await axios.post(`${API_URL}/auth/register`, {
            email: EMAIL,
            username: `admin_${TIMESTAMP}`,
            password: PASSWORD,
            fullName: 'Admin User'
        });
        console.log('Registered admin user.');
    } catch (e) {
        if (e.response && e.response.status === 400 || e.response && e.response.status === 409) {
            console.log('User already exists, logging in...');
        } else {
            console.error('Registration failed:', e.message);
            if (e.response) console.error(e.response.data);
            // Try logging in anyway
        }
    }

    // 2. Make Admin and Verify Email via SQL
    try {
        console.log('Updating user role to admin and verifying email...');
        const psqlCmd = `psql "postgresql://postgres:1234@localhost:5432/Inside_engineer1?sslmode=disable" -c "UPDATE users SET role = 'admin', email_verified = true WHERE email = '${EMAIL}';"`;
        execSync(psqlCmd, { stdio: 'inherit' });
        console.log('User role updated to admin and email verified via SQL.');
    } catch (e) {
        console.error('Failed to update role/verify email (might already be done or DB connectivity issue):', e.message);
    }

    // 3. Login
    try {
        const res = await axios.post(`${API_URL}/auth/login`, {
            email: EMAIL,
            password: PASSWORD
        });
        token = res.data.token;
        console.log('Logged in.');
    } catch (e) {
        console.error('Login failed:', e.message);
        if (e.response) console.error(e.response.data);
        process.exit(1);
    }

    // 3.4 Create Workspace
    let workspaceId;
    try {
        console.log('Creating workspace...');
        const res = await axios.post(`${API_URL}/workspaces`, {
            name: 'Test Workspace ' + TIMESTAMP,
            description: 'For testing'
        }, {
            headers: { Authorization: `Bearer ${token}` }
        });
        workspaceId = res.data.id;
        console.log('Workspace created:', workspaceId);
    } catch (e) {
        console.error('Create workspace failed:', e.message);
        if (e.response) console.error(e.response.data);
    }

    // 3.5 Create Collection
    let collectionId;
    try {
        console.log('Creating collection...');
        const res = await axios.post(`${API_URL}/collections`, {
            name: 'Test Collection ' + TIMESTAMP,
            description: 'For testing',
            workspaceId: workspaceId
        }, {
            headers: { Authorization: `Bearer ${token}` }
        });
        collectionId = res.data.id;
        console.log('Collection created:', collectionId);
    } catch (e) {
        console.error('Create collection failed:', e.message);
        if (e.response) console.error(e.response.data);
    }

    // 3. Create Dashboard
    let dashboardId;
    try {
        const res = await axios.post(`${API_URL}/dashboards`, {
            name: 'Certification Test Dashboard ' + Date.now(),
            description: 'Testing certification',
            collectionId: collectionId || 'default',
        }, {
            headers: { Authorization: `Bearer ${token}` }
        });
        dashboardId = res.data.data.id; // Fix: accesses data.id
        console.log('Dashboard created:', dashboardId);
    } catch (e) {
        console.error('Create dashboard failed:', e.message);
        if (e.response) console.error(e.response.data);
        process.exit(1);
    }

    // 4. Certify
    try {
        console.log('Certifying dashboard...');
        const res = await axios.post(`${API_URL}/dashboards/${dashboardId}/certify`, {
            status: 'verified'
        }, {
            headers: { Authorization: `Bearer ${token}` }
        });
        console.log('Certification response:', res.data);
        if (res.data.success && res.data.data.certificationStatus === 'verified') {
            console.log('SUCCESS: Dashboard certified!');
        } else {
            console.error('FAILURE: Status not verified');
            process.exit(1);
        }
    } catch (e) {
        console.error('Certification failed:', e.message);
        if (e.response) console.error(e.response.data);
        process.exit(1);
    }
}

main();
