const { AirtRuntime } = require('../src/runtime/interpreter');
const path = require('path');

async function testRuntime() {
  const runtime = new AirtRuntime();
  const mod = await runtime.runFile(path.join(__dirname, '..', 'examples/app/main.aa'));
  if (typeof mod.start !== 'function') throw new Error('start export missing');
  if (typeof mod.build !== 'function') throw new Error('build export missing');
}

async function testParseError() {
  const runtime = new AirtRuntime();
  let failed = false;
  try {
    runtime.parse('let x =', 'broken.aa');
  } catch (err) {
    failed = true;
  }
  if (!failed) throw new Error('Expected parse failure');
}

(async () => {
  await testRuntime();
  await testParseError();
  console.log('All Airt tests passed');
})();
