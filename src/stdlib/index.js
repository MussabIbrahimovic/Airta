const fs = require('fs/promises');
const path = require('path');
const { createServer } = require('./http');
const ui = require('./ui');

function createStdlib(runtime) {
  return {
    print: (...args) => console.log(...args),
    sleep: (ms) => new Promise((resolve) => setTimeout(resolve, ms)),
    env: (name, fallback = null) => process.env[name] ?? fallback,
    now: () => Date.now(),
    json: { parse: JSON.parse, stringify: (v, indent = 0) => JSON.stringify(v, null, indent) },
    path: { join: path.join, resolve: path.resolve, dirname: path.dirname, basename: path.basename },
    fs: {
      read_text: (p) => fs.readFile(p, 'utf8'),
      write_text: (p, text) => fs.writeFile(p, text, 'utf8'),
      mkdir: (p) => fs.mkdir(p, { recursive: true }),
      exists: async (p) => !!(await fs.stat(p).catch(() => null)),
    },
    http: { server: createServer },
    ui,
    runtime: {
      cwd: () => process.cwd(),
      args: () => runtime.args,
      exit: (code = 0) => process.exit(code),
    }
  };
}

module.exports = { createStdlib };
