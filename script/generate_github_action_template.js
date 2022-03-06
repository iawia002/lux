const path = require("path");
const fs = require("fs");

const extractorDir = path.join(__dirname, "..", "extractors");
const githubCIDir = path.join(__dirname, "..", ".github", "workflows");
const CITemplate = fs.readFileSync(path.join(__dirname, "github_action_template.yml"), {
  encoding: "utf-8",
});

function generateCITemplate(moduleName) {
  return CITemplate.replace(/\{\{\s*module\s*\}\}/g, moduleName);
}

const modules = fs.readdirSync(extractorDir);

const ignoreFolder = ['universal']

for (const m of modules) {
  const filepath = path.join(extractorDir, m);

  if (ignoreFolder.includes(m)) continue

  const statInfo = fs.statSync(filepath);

  if (!statInfo.isDirectory()) continue;

  fs.writeFileSync(
    path.join(githubCIDir, "stream_" + m) + ".yml",
    generateCITemplate(m)
  );
}
