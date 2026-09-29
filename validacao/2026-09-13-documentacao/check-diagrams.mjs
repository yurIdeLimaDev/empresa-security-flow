import fs from 'node:fs/promises';
import { JSDOM } from 'jsdom';

const window = new JSDOM('<!doctype html><html><body></body></html>').window;
globalThis.window = window;
globalThis.document = window.document;
const { default: mermaid } = await import('mermaid');
mermaid.initialize({ startOnLoad: false, securityLevel: 'strict' });
const input = JSON.parse(await fs.readFile(new URL('diagrams.json', import.meta.url), 'utf8'));
if (!input.length) throw new Error('No diagrams found');
const results = [];
for (const item of input) {
  try {
    await mermaid.parse(item.text);
    results.push({ source: item.source, number: item.number, status: 'passed' });
  } catch (error) {
    results.push({ source: item.source, number: item.number, status: 'failed', error: String(error) });
  }
}
const passed = results.every(item => item.status === 'passed');
const report = { status: passed ? 'passed' : 'failed', parser: 'mermaid@11.17.2',
  diagrams: results.length, results, visual_rendering_verified: false, model_calls: 0 };
await fs.writeFile(new URL('mermaid-validation.json', import.meta.url), JSON.stringify(report, null, 2)+'\n');
console.log(JSON.stringify(report));
window.close();
if (!passed) process.exitCode = 1;
