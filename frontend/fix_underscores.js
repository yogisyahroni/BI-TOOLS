const fs = require("fs");
const content = fs.readFileSync("tsc-errors.txt", "utf8");
const lines = content.split("\n");

const fixes = {};

for (const line of lines) {
  // Matches patterns like `'_CardTitle'` or `'_title'`
  const match = line.match(
    /^([a-zA-Z0-9_.\-\/\[\]]+)\((\d+),(\d+)\): error TS\d+: .*'_([a-zA-Z0-9]+)'/,
  );
  if (match) {
    const file = match[1];
    const lineNum = parseInt(match[2], 10);
    const colNum = parseInt(match[3], 10);
    const identifier = match[4];

    if (!fixes[file]) fixes[file] = [];
    fixes[file].push({ lineNum, identifier });
  }
}

for (const file of Object.keys(fixes)) {
  if (!fs.existsSync(file)) continue;
  let fileContent = fs.readFileSync(file, "utf8");

  const uniqueIdentifiers = [...new Set(fixes[file].map((f) => f.identifier))];

  for (const id of uniqueIdentifiers) {
    const regex = new RegExp(`\\b_${id}\\b`, "g");
    fileContent = fileContent.replace(regex, id);
  }

  fs.writeFileSync(file, fileContent);
  console.log(`Fixed ${uniqueIdentifiers.length} identifiers in ${file}`);
}
console.log("Done!");
