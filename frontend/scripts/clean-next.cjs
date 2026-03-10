const { existsSync, rmSync } = require("node:fs");
const { join } = require("node:path");

const nextDir = join(__dirname, "..", ".next");

if (existsSync(nextDir)) {
  rmSync(nextDir, { recursive: true, force: true });
}
