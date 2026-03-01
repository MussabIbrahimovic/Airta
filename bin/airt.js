#!/usr/bin/env node
const { runCommand, serveCommand, buildCommand, validateCommand, formatCommand } = require('../src/cli/commands');

async function main() {
  const [, , command, ...rest] = process.argv;
  try {
    if (!command || command === 'help') {
      printHelp();
      return;
    }
    if (command === 'run') {
      const [file, ...args] = rest;
      if (!file) throw new Error('Usage: airt run <file.aa> [args...]');
      await runCommand(file, args);
      return;
    }
    if (command === 'serve') {
      const file = rest[0] || 'main.aa';
      await serveCommand(file, rest.slice(1));
      return;
    }
    if (command === 'build') {
      const file = rest[0] || 'main.aa';
      const out = rest[1] || process.env.AIRT_BUILD_DIR || 'dist';
      await buildCommand(file, out);
      console.log(`Built Airt project to ${out}`);
      return;
    }
    if (command === 'validate') {
      const file = rest[0] || 'main.aa';
      await validateCommand(file);
      console.log('Validation successful');
      return;
    }
    if (command === 'format') {
      const file = rest[0] || 'main.aa';
      await formatCommand(file);
      console.log(`Formatted ${file}`);
      return;
    }
    throw new Error(`Unknown command ${command}`);
  } catch (error) {
    console.error(error.format ? error.format() : error.message);
    process.exit(1);
  }
}

function printHelp() {
  console.log(`Airt CLI\n\nCommands:\n  airt run <file.aa> [args...]\n  airt serve [file.aa]\n  airt build [file.aa] [outDir]\n  airt validate [file.aa]\n  airt format [file.aa]`);
}

main();
