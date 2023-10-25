#!/usr/bin/env node
import * as fs from "fs/promises";
import * as path from "path";

import { execSync } from 'child_process';
import { fileURLToPath } from 'url';

// Mapping from to goarch to Node's `process.arch`
var ARCH_MAPPING = {
  "386": "ia32",
  "amd64": "x64",
  "arm64": "arm64",
};

// Mapping between goos and Node's `process.platform`
var PLATFORM_MAPPING = {
  "darwin": "darwin",
  "linux": "linux",
  "windows": "win32"
};

const npmFolder = path.dirname(fileURLToPath(import.meta.url));
const NPM_DIST = path.join(npmFolder, "dist");
const RELEASES = path.join("dist", "artifacts.json");

// read the goreleaser JSON and filter down to just the binaries
const json = JSON.parse(await fs.readFile(RELEASES));
const binaries = json.filter(r => r.type === "Binary");

// clean up any previous runs
const output = execSync(`git clean -fdX ${npmFolder}`);
console.log(output.toString())

// make the dist folder
await fs.mkdir(NPM_DIST, { recursive: true })

// copy each binary into the place the NPM distribution expects it to be
await binaries.forEach(async (release) => {
  const os = PLATFORM_MAPPING[release.goos];
  const arch = ARCH_MAPPING[release.goarch];

  // use NodeJS constants for the filename, e.g. win32-x64.exe
  const distfile = `${os}-${arch}${release.extra.Ext}`;

  // copy files even if we don't use them, `package.json` uses a filtered list
  const target = path.join(NPM_DIST, distfile);
  await fs.copyFile(release.path, target);
  console.log("copied ", release.path, "to", target);
});
