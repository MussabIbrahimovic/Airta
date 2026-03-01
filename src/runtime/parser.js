class Parser {
  constructor(tokens) {
    this.tokens = tokens;
    this.pos = 0;
  }

  parseProgram() {
    const body = [];
    this.skipNewlines();
    while (!this.check('EOF')) {
      body.push(this.parseStatement());
      this.skipNewlines();
    }
    return { type: 'Program', body };
  }

  parseStatement() {
    if (this.check('IMPORT')) { this.advance(); return this.parseImport(); }
    if (this.check('LET')) { this.advance(); return this.parseLet(); }
    if (this.check('ASYNC')) { this.advance(); return this.parseFunction(true); }
    if (this.check('FUNC')) return this.parseFunction(false);
    if (this.check('COMPONENT')) { this.advance(); return this.parseComponent(); }
    if (this.check('IF')) { this.advance(); return this.parseIf(); }
    if (this.check('WHILE')) { this.advance(); return this.parseWhile(); }
    if (this.check('TRY')) { this.advance(); return this.parseTryCatch(); }
    if (this.check('RETURN')) {
      const t = this.advance();
      return { type: 'ReturnStatement', argument: this.parseMaybeExpr(), loc: t.loc };
    }
    if (this.check('THROW')) {
      const t = this.advance();
      return { type: 'ThrowStatement', argument: this.parseExpression(), loc: t.loc };
    }
    return { type: 'ExpressionStatement', expression: this.parseExpression() };
  }

  parseImport() {
    const path = this.consume('STRING', 'Expected module path').value;
    let alias = null;
    if (this.match('AS')) alias = this.consume('IDENT', 'Expected alias').value;
    return { type: 'ImportStatement', path, alias, loc: this.prev().loc };
  }

  parseLet() {
    const name = this.consume('IDENT', 'Expected variable name').value;
    this.consume('=', 'Expected = after variable name');
    const init = this.parseExpression();
    return { type: 'LetStatement', name, init, loc: this.prev().loc };
  }

  parseFunction(isAsync) {
    this.consume('FUNC', 'Expected func');
    const name = this.consume('IDENT', 'Expected function name').value;
    const params = this.parseParams();
    this.skipNewlines();
    const body = this.parseBlock(['END']);
    this.consume('END', 'Expected end for function');
    return { type: 'FunctionDeclaration', name, params, body, async: isAsync, loc: this.prev().loc };
  }

  parseComponent() {
    const name = this.consume('IDENT', 'Expected component name').value;
    const params = this.parseParams();
    this.skipNewlines();
    const body = this.parseBlock(['END']);
    this.consume('END', 'Expected end for component');
    return { type: 'ComponentDeclaration', name, params, body, loc: this.prev().loc };
  }

  parseIf() {
    const test = this.parseExpression();
    this.skipNewlines();
    const consequent = this.parseBlock(['ELSE', 'END']);
    let alternate = [];
    if (this.match('ELSE')) {
      this.skipNewlines();
      alternate = this.parseBlock(['END']);
    }
    this.consume('END', 'Expected end for if');
    return { type: 'IfStatement', test, consequent, alternate, loc: this.prev().loc };
  }

  parseWhile() {
    const test = this.parseExpression();
    this.skipNewlines();
    const body = this.parseBlock(['END']);
    this.consume('END', 'Expected end for while');
    return { type: 'WhileStatement', test, body, loc: this.prev().loc };
  }

  parseTryCatch() {
    this.skipNewlines();
    const block = this.parseBlock(['CATCH']);
    this.consume('CATCH', 'Expected catch block');
    const name = this.consume('IDENT', 'Expected catch variable').value;
    this.skipNewlines();
    const handler = this.parseBlock(['END']);
    this.consume('END', 'Expected end for try/catch');
    return { type: 'TryCatchStatement', block, name, handler, loc: this.prev().loc };
  }

  parseBlock(stoppers = ['END']) {
    const stmts = [];
    this.skipNewlines();
    while (!this.check('EOF') && !stoppers.some((s) => this.check(s))) {
      stmts.push(this.parseStatement());
      this.skipNewlines();
    }
    return stmts;
  }

  parseParams() {
    const params = [];
    this.consume('(', 'Expected (');
    while (!this.check(')')) {
      params.push(this.consume('IDENT', 'Expected parameter').value);
      if (!this.match(',')) break;
    }
    this.consume(')', 'Expected )');
    return params;
  }

  parseMaybeExpr() {
    if (this.check('NEWLINE') || this.check('EOF') || this.check('END')) return null;
    return this.parseExpression();
  }

  parseExpression(precedence = 0) {
    let left = this.parsePrefix();
    while (true) {
      const op = this.peek();
      const prec = this.getPrecedence(op.type);
      if (prec <= precedence) break;
      this.advance();
      left = { type: 'BinaryExpression', operator: op.type, left, right: this.parseExpression(prec), loc: op.loc };
    }
    return left;
  }

  parsePrefix() {
    if (this.match('NUMBER')) return { type: 'Literal', value: this.prev().value, loc: this.prev().loc };
    if (this.match('STRING')) return { type: 'Literal', value: this.prev().value, loc: this.prev().loc };
    if (this.match('TRUE')) return { type: 'Literal', value: true, loc: this.prev().loc };
    if (this.match('FALSE')) return { type: 'Literal', value: false, loc: this.prev().loc };
    if (this.match('NULL')) return { type: 'Literal', value: null, loc: this.prev().loc };
    if (this.match('NOT') || this.match('-')) {
      const op = this.prev();
      return { type: 'UnaryExpression', operator: op.type, argument: this.parseExpression(7), loc: op.loc };
    }
    if (this.match('AWAIT')) {
      const token = this.prev();
      return { type: 'AwaitExpression', argument: this.parseExpression(7), loc: token.loc };
    }
    if (this.match('(')) {
      const expr = this.parseExpression();
      this.consume(')', 'Expected )');
      return this.parsePostfix(expr);
    }
    if (this.match('[')) {
      const elements = [];
      while (!this.check(']')) {
        elements.push(this.parseExpression());
        if (!this.match(',')) break;
      }
      this.consume(']', 'Expected ]');
      return { type: 'ArrayExpression', elements, loc: this.prev().loc };
    }
    if (this.match('{')) {
      const properties = [];
      while (!this.check('}')) {
        const key = this.consume('IDENT', 'Expected object key').value;
        this.consume(':', 'Expected : after key');
        const value = this.parseExpression();
        properties.push({ key, value });
        if (!this.match(',')) break;
      }
      this.consume('}', 'Expected }');
      return { type: 'ObjectExpression', properties, loc: this.prev().loc };
    }
    if (this.match('IDENT')) {
      const expr = { type: 'Identifier', name: this.prev().value, loc: this.prev().loc };
      return this.parsePostfix(expr);
    }
    throw new Error(`Unexpected token ${this.peek().type} at ${this.peek().loc.file}:${this.peek().loc.line}:${this.peek().loc.col}`);
  }

  parsePostfix(base) {
    let expr = base;
    while (true) {
      if (this.match('(')) {
        const args = [];
        while (!this.check(')')) {
          args.push(this.parseExpression());
          if (!this.match(',')) break;
        }
        this.consume(')', 'Expected )');
        expr = { type: 'CallExpression', callee: expr, args, loc: this.prev().loc };
        continue;
      }
      if (this.match('.')) {
        const property = this.consume('IDENT', 'Expected property name').value;
        expr = { type: 'MemberExpression', object: expr, property, loc: this.prev().loc };
        continue;
      }
      if (this.match('=')) {
        const value = this.parseExpression();
        expr = { type: 'AssignmentExpression', left: expr, right: value, loc: this.prev().loc };
      }
      break;
    }
    return expr;
  }

  getPrecedence(op) {
    return { or: 1, and: 2, '==': 3, '!=': 3, '<': 4, '>': 4, '<=': 4, '>=': 4, '+': 5, '-': 5, '*': 6, '/': 6, '%': 6 }[op] || 0;
  }

  check(type) { return this.peek().type === type; }
  peek() { return this.tokens[this.pos]; }
  prev() { return this.tokens[this.pos - 1]; }
  advance() { return this.tokens[this.pos++]; }
  match(type) { if (this.check(type)) { this.advance(); return true; } return false; }
  consume(type, message) { if (this.check(type)) return this.advance(); throw new Error(`${message} near ${this.peek().loc.file}:${this.peek().loc.line}:${this.peek().loc.col}`); }
  skipNewlines() { while (this.match('NEWLINE')); }
}

module.exports = { Parser };
