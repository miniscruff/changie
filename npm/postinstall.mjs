#!/usr/bin/env node
import * as fs from "fs/promises";
import * as path from "path";

// A list of files distributed through NPM, this should mirror the package.json file list
export const executables = {
  "darwin-arm64": "changie",
  "darwin-x64": "changie",
  "linux-arm64": "changie",
  "linux-x64": "changie",
  "win32-ia32.exe": "changie.exe",
  "win32-x64.exe": "changie.exe"
};

const copyChangie = async (filename, changie) => {
  const source = path.join(DIST, filename);
  const stat = await fs.stat(source)
  if (stat.isFile()) {
    await fs.copyFile(source, changie)
  } else {
    console.warn(`Unable to find changie ${source} in NPM package`)
  }
}

const DIST = "npm/dist";
const FALLBACK = "npm/changie.js";

const ext = process.platform === 'win32' ? '.exe' : '';
const target = `${process.platform}-${process.arch}${ext}`

if (executables[target]) {
  copyChangie(target, executables[target]);
} else {
  // use `changie.js` which will throw an error when run
  await fs.copyFile(FALLBACK, "changie");
}
