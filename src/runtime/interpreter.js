const fs = require('fs/promises');
const path = require('path');
const { Lexer } = require('./lexer');
const { Parser } = require('./parser');
const { Environment } = require('./environment');
const { AirtError } = require('./errors');
const { createStdlib } = require('../stdlib');

class ReturnSignal { constructor(value) { this.value = value; } }

class AirtRuntime {
  constructor(options = {}) {
    this.moduleCache = new Map();
    this.args = options.args || [];
    this.rootEnv = new Environment();
    const std = createStdlib(this);
    for (const [k, v] of Object.entries(std)) this.rootEnv.define(k, v);
  }

  async runFile(entryFile) {
    const full = path.resolve(entryFile);
    const exports = await this.executeModule(full);
    return exports;
  }

  async executeModule(filePath) {
    const resolved = path.resolve(filePath);
    if (this.moduleCache.has(resolved)) return this.moduleCache.get(resolved);
    const source = await fs.readFile(resolved, 'utf8');
    const ast = this.parse(source, resolved);
    const moduleEnv = new Environment(this.rootEnv);
    const moduleExports = {};
    moduleEnv.define('exports', moduleExports);
    this.moduleCache.set(resolved, moduleExports);
    await this.evalBlock(ast.body, moduleEnv);
    return moduleExports;
  }

  parse(source, file) {
    const tokens = new Lexer(source, file).tokenize();
    return new Parser(tokens).parseProgram();
  }

  async evalBlock(statements, env) {
    for (const stmt of statements) {
      const out = await this.evalStatement(stmt, env);
      if (out instanceof ReturnSignal) return out;
    }
    return null;
  }

  async evalStatement(node, env) {
    switch (node.type) {
      case 'ImportStatement': {
        const modulePath = path.resolve(path.dirname(node.loc.file), node.path);
        const exports = await this.executeModule(modulePath);
        if (node.alias) env.define(node.alias, exports);
        return null;
      }
      case 'LetStatement':
        return env.define(node.name, await this.evalExpression(node.init, env));
      case 'FunctionDeclaration': {
        const fn = this.createFunction(node, env, node.name);
        env.define(node.name, fn);
        return null;
      }
      case 'ComponentDeclaration': {
        const fn = this.createFunction({ ...node, type: 'FunctionDeclaration' }, env, node.name);
        fn.__airtComponent = true;
        env.define(node.name, fn);
        return null;
      }
      case 'ExpressionStatement':
        return await this.evalExpression(node.expression, env);
      case 'IfStatement': {
        if (this.truthy(await this.evalExpression(node.test, env))) return this.evalBlock(node.consequent, new Environment(env));
        return this.evalBlock(node.alternate, new Environment(env));
      }
      case 'WhileStatement': {
        while (this.truthy(await this.evalExpression(node.test, env))) {
          const out = await this.evalBlock(node.body, new Environment(env));
          if (out instanceof ReturnSignal) return out;
        }
        return null;
      }
      case 'ReturnStatement':
        return new ReturnSignal(node.argument ? await this.evalExpression(node.argument, env) : null);
      case 'ThrowStatement': {
        const err = await this.evalExpression(node.argument, env);
        throw new AirtError(typeof err === 'string' ? err : JSON.stringify(err), node);
      }
      case 'TryCatchStatement': {
        try {
          return await this.evalBlock(node.block, new Environment(env));
        } catch (err) {
          const scope = new Environment(env);
          scope.define(node.name, err.message || String(err));
          return await this.evalBlock(node.handler, scope);
        }
      }
      default:
        throw new AirtError(`Unsupported statement: ${node.type}`, node);
    }
  }

  createFunction(node, env, name) {
    const runtime = this;
    const fn = async function (...args) {
      const scope = new Environment(env);
      node.params.forEach((p, idx) => scope.define(p, args[idx] ?? null));
      const out = await runtime.evalBlock(node.body, scope);
      return out instanceof ReturnSignal ? out.value : null;
    };
    Object.defineProperty(fn, 'name', { value: name || 'anonymous' });
    return fn;
  }

  async evalExpression(node, env) {
    switch (node.type) {
      case 'Literal': return node.value;
      case 'Identifier': return env.get(node.name, node);
      case 'ArrayExpression': return Promise.all(node.elements.map((el) => this.evalExpression(el, env)));
      case 'ObjectExpression': {
        const obj = {};
        for (const prop of node.properties) obj[prop.key] = await this.evalExpression(prop.value, env);
        return obj;
      }
      case 'UnaryExpression': {
        const v = await this.evalExpression(node.argument, env);
        if (node.operator === '-') return -v;
        if (node.operator === 'NOT') return !this.truthy(v);
        throw new AirtError(`Unknown unary operator ${node.operator}`, node);
      }
      case 'BinaryExpression': return this.evalBinary(node, env);
      case 'CallExpression': {
        const callee = await this.evalExpression(node.callee, env);
        const args = [];
        for (const arg of node.args) args.push(await this.evalExpression(arg, env));
        if (typeof callee !== 'function') throw new AirtError('Attempted to call a non-function value', node);
        return await callee(...args);
      }
      case 'MemberExpression': {
        const obj = await this.evalExpression(node.object, env);
        return obj?.[node.property];
      }
      case 'AssignmentExpression': {
        if (node.left.type === 'Identifier') return env.assign(node.left.name, await this.evalExpression(node.right, env), node);
        if (node.left.type === 'MemberExpression') {
          const obj = await this.evalExpression(node.left.object, env);
          obj[node.left.property] = await this.evalExpression(node.right, env);
          return obj[node.left.property];
        }
        throw new AirtError('Invalid assignment target', node);
      }
      case 'AwaitExpression':
        return await this.evalExpression(node.argument, env);
      default:
        throw new AirtError(`Unsupported expression: ${node.type}`, node);
    }
  }

  async evalBinary(node, env) {
    const left = await this.evalExpression(node.left, env);
    if (node.operator === 'and') return this.truthy(left) ? this.evalExpression(node.right, env) : left;
    if (node.operator === 'or') return this.truthy(left) ? left : this.evalExpression(node.right, env);
    const right = await this.evalExpression(node.right, env);
    switch (node.operator) {
      case '+': return left + right;
      case '-': return left - right;
      case '*': return left * right;
      case '/': return left / right;
      case '%': return left % right;
      case '==': return left === right;
      case '!=': return left !== right;
      case '<': return left < right;
      case '<=': return left <= right;
      case '>': return left > right;
      case '>=': return left >= right;
      default: throw new AirtError(`Unknown operator ${node.operator}`, node);
    }
  }

  truthy(value) { return !!value; }
}

module.exports = { AirtRuntime };
