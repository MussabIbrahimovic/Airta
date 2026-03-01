const KEYWORDS = new Set([
  'let', 'func', 'async', 'return', 'if', 'else', 'while', 'for', 'in', 'true', 'false', 'null', 'import', 'as', 'throw', 'try', 'catch', 'component', 'await', 'and', 'or', 'not', 'end'
]);

class Lexer {
  constructor(input, file = '<memory>') {
    this.input = input.replace(/\r\n/g, '\n');
    this.file = file;
    this.pos = 0;
    this.line = 1;
    this.col = 1;
    this.tokens = [];
  }

  tokenize() {
    while (!this.eof()) {
      const ch = this.peek();
      if (ch === ' ' || ch === '\t') {
        this.advance();
        continue;
      }
      if (ch === '\n') {
        this.push('NEWLINE', '\n');
        this.advanceLine();
        continue;
      }
      if (ch === '#') {
        while (!this.eof() && this.peek() !== '\n') this.advance();
        continue;
      }
      if (/[0-9]/.test(ch)) {
        this.tokenizeNumber();
        continue;
      }
      if (/[A-Za-z_]/.test(ch)) {
        this.tokenizeIdentifier();
        continue;
      }
      if (ch === '"' || ch === "'") {
        this.tokenizeString(ch);
        continue;
      }
      this.tokenizeSymbol();
    }
    this.tokens.push(this.mk('EOF', null));
    return this.tokens;
  }

  tokenizeNumber() {
    const start = this.loc();
    let text = '';
    while (!this.eof() && /[0-9.]/.test(this.peek())) text += this.advance();
    this.tokens.push({ type: 'NUMBER', value: Number(text), raw: text, loc: start });
  }

  tokenizeIdentifier() {
    const start = this.loc();
    let text = '';
    while (!this.eof() && /[A-Za-z0-9_]/.test(this.peek())) text += this.advance();
    if (KEYWORDS.has(text)) {
      this.tokens.push({ type: text.toUpperCase(), value: text, loc: start });
    } else {
      this.tokens.push({ type: 'IDENT', value: text, loc: start });
    }
  }

  tokenizeString(quote) {
    const start = this.loc();
    this.advance();
    let out = '';
    while (!this.eof() && this.peek() !== quote) {
      if (this.peek() === '\\') {
        this.advance();
        const esc = this.advance();
        const map = { n: '\n', t: '\t', r: '\r', '"': '"', "'": "'", '\\': '\\' };
        out += map[esc] ?? esc;
      } else {
        out += this.advance();
      }
    }
    this.advance();
    this.tokens.push({ type: 'STRING', value: out, loc: start });
  }

  tokenizeSymbol() {
    const start = this.loc();
    const two = this.input.slice(this.pos, this.pos + 2);
    const pairs = ['==', '!=', '<=', '>=', '=>'];
    if (pairs.includes(two)) {
      this.advance(); this.advance();
      this.tokens.push({ type: two, value: two, loc: start });
      return;
    }
    const one = this.advance();
    const singles = '(){}[],:.+-*/%<>=.';
    if (singles.includes(one)) {
      this.tokens.push({ type: one, value: one, loc: start });
      return;
    }
    throw new Error(`Unexpected token ${one} at ${this.file}:${start.line}:${start.col}`);
  }

  push(type, value) { this.tokens.push(this.mk(type, value)); }
  mk(type, value) { return { type, value, loc: this.loc() }; }
  loc() { return { file: this.file, line: this.line, col: this.col }; }
  eof() { return this.pos >= this.input.length; }
  peek() { return this.input[this.pos]; }
  advance() { const c = this.input[this.pos++]; this.col += 1; return c; }
  advanceLine() { this.pos += 1; this.line += 1; this.col = 1; }
}

module.exports = { Lexer };
