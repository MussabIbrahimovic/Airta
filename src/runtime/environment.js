const { AirtError } = require('./errors');

class Environment {
  constructor(parent = null) {
    this.parent = parent;
    this.values = new Map();
  }

  define(name, value) {
    this.values.set(name, value);
    return value;
  }

  assign(name, value, node) {
    if (this.values.has(name)) {
      this.values.set(name, value);
      return value;
    }
    if (this.parent) return this.parent.assign(name, value, node);
    throw new AirtError(`Undefined variable '${name}'`, node);
  }

  get(name, node) {
    if (this.values.has(name)) return this.values.get(name);
    if (this.parent) return this.parent.get(name, node);
    throw new AirtError(`Undefined variable '${name}'`, node);
  }
}

module.exports = { Environment };
