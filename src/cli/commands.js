const fs = require('fs/promises');
const path = require('path');
const { AirtRuntime } = require('../runtime/interpreter');

async function runCommand(file, args) {
  const runtime = new AirtRuntime({ args });
  return runtime.runFile(file);
}

async function serveCommand(file, args) {
  const exports = await runCommand(file, args);
  if (typeof exports.start !== 'function') throw new Error('serve expects module to export start()');
  await exports.start();
}

async function buildCommand(file, outDir = 'dist') {
  const runtime = new AirtRuntime();
  const exports = await runtime.runFile(file);
  await fs.mkdir(outDir, { recursive: true });
  if (typeof exports.build === 'function') {
    await exports.build(outDir);
  } else {
    const source = await fs.readFile(file, 'utf8');
    await fs.writeFile(path.join(outDir, 'bundle.aa'), source, 'utf8');
  }
  return outDir;
}

async function validateCommand(file) {
  const source = await fs.readFile(file, 'utf8');
  const runtime = new AirtRuntime();
  runtime.parse(source, path.resolve(file));
  return true;
}

async function formatCommand(file) {
  const source = await fs.readFile(file, 'utf8');
  const lines = source.split('\n').map((line) => line.replace(/\t/g, '  ').replace(/\s+$/g, ''));
  const formatted = lines.join('\n').replace(/\n{3,}/g, '\n\n');
  await fs.writeFile(file, formatted, 'utf8');
  return formatted;
}

module.exports = { runCommand, serveCommand, buildCommand, validateCommand, formatCommand };
