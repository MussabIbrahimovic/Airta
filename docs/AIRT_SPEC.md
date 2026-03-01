# Airt

## 1) High-level vision of Airt
Airt is a unified .aa programming language for application logic, web APIs, UI composition, static generation, and automation scripts using one runtime and one syntax.

## 2) Language philosophy and design goals
- One language for backend + frontend concerns.
- Human-readable syntax with Python-like blocks terminated by `end`.
- First-class async/await and deterministic event-loop behavior.
- Native UI tree and style objects in language values.
- Real runtime + CLI without host-language syntax exposure.

## 3) Full syntax specification
### Lexical
- Comments: `# ...`
- Strings: `'...'` and `"..."`
- Numbers: decimal numeric literals
- Keywords: `let func async return if else while true false null import as throw try catch component await and or not end`

### Statements
- Variable: `let name = expression`
- Function:
  ```
  func name(a, b)
    ...
    return value
  end
  ```
- Async function: `async func ... end`
- Component declaration: `component Name(props) ... end`
- Conditionals:
  ```
  if cond
    ...
  else
    ...
  end
  ```
- Loop:
  ```
  while cond
    ...
  end
  ```
- Imports: `import "./module.aa" as alias`
- Exceptions:
  ```
  try
    ...
  catch err
    ...
  end
  ```
- Throws: `throw expression`

### Expressions
- Literals: numbers, strings, booleans, null
- Arrays: `[a, b]`
- Objects: `{ key: value }`
- Unary: `-x`, `not x`, `await expr`
- Binary: `+ - * / % == != < <= > >= and or`
- Calls and member access: `f(x)`, `obj.method`
- Assignment: `name = value`, `obj.prop = value`

## 4) Execution model
1. `.aa` source is tokenized by `Lexer`.
2. Tokens are parsed into AST by `Parser`.
3. AST is executed in lexical environments by `AirtRuntime`.
4. Module imports resolve relative file paths and are cached.
5. Async calls are Promise-based and use explicit `await`.

## 5) Runtime architecture
- `src/runtime/lexer.js`: lexical analysis.
- `src/runtime/parser.js`: recursive descent + precedence parser.
- `src/runtime/environment.js`: lexical scope chain.
- `src/runtime/interpreter.js`: evaluator, function closures, modules.
- `src/runtime/errors.js`: AirtError with source location and stack formatting.
- `src/stdlib/*`: standard library modules.

## 6) CLI design
`airt` supports:
- `airt run main.aa`
- `airt serve main.aa`
- `airt build main.aa dist`
- `airt validate main.aa`
- `airt format main.aa`

## 7) Standard library overview
- Core: `print`, `sleep`, `env`, `now`
- JSON: `json.parse`, `json.stringify`
- File system: `fs.read_text`, `fs.write_text`, `fs.mkdir`, `fs.exists`
- Paths: `path.join`, `path.resolve`, `path.dirname`, `path.basename`
- HTTP: `http.server()` with `use/get/post/listen`
- UI: `ui.element`, `ui.text`, `ui.renderToHtml`
- Runtime: `runtime.cwd`, `runtime.args`, `runtime.exit`

## 8) Complete folder and file structure
See repository tree.

## 9) Complete source code for all files
Implemented in this repository and installable via npm as `airt-lang`.

## 10) Installation instructions
```
npm install
npm link
airt help
```

## 11) Usage examples written in .aa
See `examples/app/main.aa` and `examples/app/components.aa`.

## 12) Error handling behavior
- Syntax errors include file/line/column.
- Runtime errors are raised as `AirtError` with source location.
- CLI prints formatted error and exits with code 1.

## 13) Future extension points
- Bytecode compiler and VM backend.
- Type checker and LSP.
- Router params and request body parser.
- JSX-like sugar for UI nodes.
- Package registry and module version solver.
