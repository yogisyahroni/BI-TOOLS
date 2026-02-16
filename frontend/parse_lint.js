const fs = require('fs');
try {
    const content = fs.readFileSync('lint_output_final.txt', 'utf8');
    const lines = content.split('\n');
    let currentFile = '';
    for (const line of lines) {
        if (line.startsWith('./')) {
            currentFile = line.trim();
        } else if (line.includes('Error:')) {
            console.log(`FILE: ${currentFile} || MSG: ${line.trim()}`);
        }
    }
} catch (e) {
    console.error(e);
}
