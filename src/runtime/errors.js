class AirtError extends Error {
  constructor(message, node, frames = []) {
    super(message);
    this.name = 'AirtError';
    this.node = node || null;
    this.frames = frames;
  }

  format() {
    const location = this.node?.loc ? ` (${this.node.loc.file}:${this.node.loc.line}:${this.node.loc.col})` : '';
    const stack = this.frames.length
      ? `\nStack trace:\n${this.frames.map((f) => `  at ${f.fn} (${f.file}:${f.line}:${f.col})`).join('\n')}`
      : '';
    return `${this.name}: ${this.message}${location}${stack}`;
  }
}

module.exports = { AirtError };
