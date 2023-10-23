#!/usr/bin/env node
import * as fs from "fs/promises";
import * as path from "path";

const DIST = "dist";

const copyChangie = async (filename, changie) => {
  const source = path.join(DIST, filename);
  const stat = await fs.stat(source)
  if (stat.isFile()) {
    await fs.copyFile(source, changie)
  } else {
    console.warn(`Unable to find changie ${source} in NPM package`)
  }
}

switch (process.platform) {
  case "win32":
    copyChangie(`${process.platform}-${process.arch}.exe`, "changie.exe");
  case "darwin":
  case "linux":
    copyChangie(`${process.platform}-${process.arch}`, "changie");
  default:
    // leave `changie.js` in place which will throw an error when used
}
