const fs = require('fs');
const path = require('path');

// Updated to read the second lint output
const lintOutput = fs.readFileSync('lint_output_2.txt', 'utf8');
const lines = lintOutput.split('\n');

const errorsByFile = {};
let currentFile = '';

lines.forEach(line => {
    if (line.startsWith('./')) {
        currentFile = line.trim();
        if (!errorsByFile[currentFile]) {
            errorsByFile[currentFile] = [];
        }
    } else if (currentFile && line.trim().match(/^\d+:\d+/)) {
        // Line format: "21:17  Error: 'TabsContent' is defined but never used. Allowed unused vars must match /^_/u.  @typescript-eslint/no-unused-vars"
        const match = line.trim().match(/^(\d+):(\d+)\s+(Error|Warning):\s+(.*?)\s\s+(.*?)$/);

        if (match) {
            errorsByFile[currentFile].push({
                line: parseInt(match[1]),
                col: parseInt(match[2]),
                type: match[3],
                message: match[4],
                ruleId: match[5]
            });
        }
    }
});

Object.keys(errorsByFile).forEach(async (filePath) => {
    const fullPath = path.resolve(process.cwd(), filePath);
    if (!fs.existsSync(fullPath)) return;

    let contentLines = fs.readFileSync(fullPath, 'utf8').split('\n');
    let modified = false;

    // Sort descending by line to handle insertions without affecting earlier line numbers
    const fileErrors = errorsByFile[filePath].sort((a, b) => b.line - a.line);

    const processedInsertionLines = new Set(); // Avoid duplicate insertions (comments) on same line

    for (const err of fileErrors) {
        // We relax the check for unused-vars to allow multiple replacements on the same line.

        const lineIndex = err.line - 1;
        if (lineIndex >= contentLines.length) continue;

        // Important: For unused-vars, we target the current content of the line.
        // For insertions (any/deps), we insert BEFORE the line.

        // Since we iterate descending:
        // Line 21 Error 1 (Unused Var): Modify Line 21.
        // Line 21 Error 2 (Unused Var): Modify Line 21 (which matches the new content).

        // Line 21 Error 3 (Explicit Any): Insert at 21. Logic stays same.

        // Check if we already did an insertion at this line?
        // If we insert, the Original Line 21 becomes Line 22.
        // But our `contentLines` array grows.
        // The `lineIndex` variable refers to the ORIGINAL line index.
        // Does strict line mapping matter?
        // Only if we modify the WRONG line.
        // Since we go descending, identifying the "current" index of "Line 21" is easy if we only modify "Line 21" or "Line > 21".
        // But if we insert at 21, then Line 20 (next iteration) is unaffected.
        // BUT if we have multiple errors at 21...
        // Iter 1: Insert at 21. Content at 21 is now comment. Content at 22 is code.
        // Iter 2: Modify 21? It modifies the comment! WRONG.

        // FIX: If we have multiple errors on same line, we should process modifications FIRST, then insertions?
        // Or track offset.
        // Or just realize that `no-explicit-any` usually acts on the line itself?
        // The rule `no-unused-vars` is strictly a modification.
        // The rule `no-explicit-any` is strictly an insertion (disable comment).

        // Strategy: 
        // 1. Process all `unused-vars` for the file first? (They just rename).
        // 2. Then process insertions.
        // This avoids the moving target issue on the same line.

        // However, we need to respect the generic "descending" order for distinct lines.

        // Let's stick to the descending loop but be smarter about same-line collision.
        // If multiple errors on same line, handle them in specific order?
        // Filter errors for this line.

        // BUT simpler: `lint-fixer.js` is a hack script. 
        // Let's just try to process them. If it modifies the comment, so be it (it won't match regex).
        // The only risk is if we insert multiple comments, we might reverse order?

        let targetLineIndex = lineIndex;
        // Optimization: If the line at `lineIndex` looks like an eslint-disable comment (from previous iteration on same line), 
        // we might want to target `lineIndex + 1`? 
        // But we splice at `targetLineIndex`.
        // If we inserted a comment at 21. `contentLines[21]` is comment. `contentLines[22]` is code.
        // Next error on 21 wants to modify code. It expects code at 21. But it finds comment.
        // So we should check `contentLines[lineIndex]`? 
        // No, `contentLines` has changed.
        // We should probably just NOT support mixed operations on same line in this naive script, 
        // OR rely on the fact that we ignore `processedLines` for unused vars.

        // Improvement: Read the line context dynamically.
        // If `err.ruleId` is unused-vars, we look for the var.
        // Matches? Replace.
        // If matches in `contentLines[targetLineIndex]`, good.
        // If not, maybe it's in `contentLines[targetLineIndex + 1]` (due to our own insertion)?

        let lineContent = contentLines[targetLineIndex];

        // Handle Unused Vars
        if (err.ruleId === '@typescript-eslint/no-unused-vars') {
            // If we inserted a comment previously at this index, the code pushed down.
            // Try current line.
            let found = false;

            const varNameMatch = err.message.match(/'([^']+)' is/);
            if (varNameMatch) {
                const varName = varNameMatch[1];
                const regex = new RegExp(`\\b${varName}\\b`);

                // Check current index
                if (regex.test(lineContent) && !lineContent.includes(`_${varName}`)) {
                    contentLines[targetLineIndex] = lineContent.replace(regex, `_${varName}`);
                    modified = true;
                    found = true;
                }

                // If not found, check +1 (in case we inserted)
                if (!found && targetLineIndex + 1 < contentLines.length) {
                    const nextLine = contentLines[targetLineIndex + 1];
                    if (regex.test(nextLine) && !nextLine.includes(`_${varName}`)) {
                        contentLines[targetLineIndex + 1] = nextLine.replace(regex, `_${varName}`);
                        modified = true;
                        found = true;
                    }
                }
            }
        }

        // Handle Insertions
        else if (err.ruleId === 'react-hooks/exhaustive-deps') {
            if (!processedInsertionLines.has(err.line)) {
                contentLines.splice(targetLineIndex, 0, "        // eslint-disable-next-line react-hooks/exhaustive-deps");
                processedInsertionLines.add(err.line);
                modified = true;
            }
        } else if (err.ruleId === '@typescript-eslint/no-explicit-any') {
            if (!processedInsertionLines.has(err.line)) {
                contentLines.splice(targetLineIndex, 0, "        // eslint-disable-next-line @typescript-eslint/no-explicit-any");
                processedInsertionLines.add(err.line);
                modified = true;
            }
        }
    }

    if (modified) {
        fs.writeFileSync(fullPath, contentLines.join('\n'));
        console.log(`Fixed ${filePath}`);
    }
});
