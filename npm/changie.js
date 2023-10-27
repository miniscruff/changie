#!/usr/bin/env node
const fs = require('node:fs');
const path = require('node:path');
const { spawn } = require('node:child_process')

// Platform distributed through NPM, this should mirror the package.json file list
const executables = {
  'darwin-arm64': true,
  'darwin-x64': true,
  'linux-arm64': true,
  'linux-x64': true,
  'win32-ia32': true,
  'win32-x64': true
};

const DIST = path.join(__dirname, 'dist');
const target = `${process.platform}-${process.arch}`

const runChangie = (filename) => {
  const ext = process.platform === 'win32' ? '.exe' : '';
  const executable = path.join(DIST, filename + ext);
  const stat = fs.statSync(executable)
  if (stat.isFile()) {
    const child = spawn(executable, process.argv.slice(2));
    child.stdout.pipe(process.stdout);
    child.stderr.pipe(process.stderr);
    child.on('close', (code) => {
      process.exit(code);
    });
  } else {
    throw new Error(`Unable to find changie ${executable} in NPM package`)
  }
}

if (executables[target]) {
  runChangie(target);
} else {
  throw new Error(`Unsupported platform for Changie: ${target}`);
}
