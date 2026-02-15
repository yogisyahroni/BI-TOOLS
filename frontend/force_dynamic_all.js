
const fs = require('fs');
const path = require('path');

function walkDir(dir, callback) {
    fs.readdirSync(dir).forEach(f => {
        let dirPath = path.join(dir, f);
        let isDirectory = fs.statSync(dirPath).isDirectory();
        isDirectory ? walkDir(dirPath, callback) : callback(path.join(dir, f));
    });
}

const appDir = path.join(__dirname, 'app');

console.log(`Scanning ${appDir} for page.tsx files...`);

walkDir(appDir, (filePath) => {
    const filename = path.basename(filePath);
    if (['page.tsx', 'route.ts', 'not-found.tsx'].includes(filename)) {
        let content = fs.readFileSync(filePath, 'utf8');

        // Check if already has dynamic export
        if (content.includes('export const dynamic')) {
            console.log(`Skipping (already has dynamic export): ${filePath}`);
            return;
        }

        // Add correct header
        // If 'use client' is present, add after it.
        // If not, add at top (and adding 'use client' might be risky for server components that expect headers/cookies, but force-dynamic works for both client and server components).
        // Wait, force-dynamic works on Server Components too.

        let newContent = content;
        if (content.includes("'use client'")) {
            newContent = content.replace(/'use client';?/, "'use client';\n\nexport const dynamic = 'force-dynamic';");
        } else if (content.includes('"use client"')) {
            newContent = content.replace(/"use client";?/, '"use client";\n\nexport const dynamic = \'force-dynamic\';');
        } else {
            // Server component or no directive
            newContent = `export const dynamic = 'force-dynamic';\n\n${content}`;
        }

        fs.writeFileSync(filePath, newContent);
        console.log(`Updated: ${filePath}`);
    }
});
