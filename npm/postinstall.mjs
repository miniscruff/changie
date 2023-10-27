#!/usr/bin/env node
import * as fs from 'node:fs/promises';
import * as path from 'node:path';

// A list of files distributed through NPM, this should mirror the package.json file list
export const executables = {
  'darwin-arm64': 'changie',
  'darwin-x64': 'changie',
  'linux-arm64': 'changie',
  'linux-x64': 'changie',
  'win32-ia32.exe': 'changie.exe',
  'win32-x64.exe': 'changie.exe'
};

const NPM = 'npm'
const DIST = path.join(NPM, 'dist');
const FALLBACK = path.join(NPM, 'changie');
const ext = process.platform === 'win32' ? '.exe' : '';
const target = `${process.platform}-${process.arch}${ext}`

const copyChangie = async (filename, changie) => {
  const source = path.join(DIST, filename);
  const stat = await fs.stat(source)
  if (stat.isFile()) {
    // remove the fallback, just to be safe, also to not confuse windows
    await fs.unlink(FALLBACK);
    const target = path.join(NPM, changie);
    await fs.copyFile(source, target);
  } else {
    console.warn(`Unable to find changie ${source} in NPM package`)
  }
}

if (executables[target]) {
  await copyChangie(target, executables[target]);
} else {
  // the fallback is already in place, a script that throws an error
}
