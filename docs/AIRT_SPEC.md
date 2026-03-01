# Airt

## 1) High-level vision of Airt
Airt is a unified, standalone language for backend logic, UI composition, HTTP APIs, static builds, and CLI automation using only `.aa` code.

## 2) Language philosophy and design goals
- Single language for app/server/UI/build tasks.
- Readable block syntax with `end` terminators.
- Native async-style `await` semantics.
- Built-in UI tree and style model.
- No host-language syntax in `.aa` programs.
- Runtime and CLI are native binaries.

## 3) Full syntax specification
Keywords: `let func async return if else while true false null import as throw try catch component await and or not end`.

Statements:
- `import "./module.aa" as alias`
- `let name = expr`
- `func name(a, b) ... end`
- `async func name(...) ... end`
- `component Name(props) ... end`
- `if cond ... else ... end`
- `while cond ... end`
- `try ... catch err ... end`
- `return expr`
- `throw expr`

Expressions:
- literals, arrays, objects
- unary: `-`, `not`, `await`
- binary: `+ - * / % == != < <= > >= and or`
- calls, member access, assignments

## 4) Execution model
1. Lexer tokenizes `.aa` source.
2. Parser builds AST.
3. Runtime evaluates AST in lexical environments.
4. Imports are resolved relative to importing module and cached.
5. Stdlib is injected as native runtime values.

## 5) Runtime architecture
- `internal/airt/lexer`: tokenizer
- `internal/airt/parser`: recursive descent parser
- `internal/airt/ast`: AST nodes
- `internal/airt/runtime`: evaluator/module loader/scope engine
- `internal/airt/stdlib`: HTTP/UI/fs/path/env/json/runtime built-ins

## 6) CLI design
`airt run`, `airt serve`, `airt build`, `airt validate`, `airt format`

## 7) Standard library overview
- Core: `print`, `now`, `sleep`, `env`
- FS/path: `fs.read_text`, `fs.write_text`, `fs.mkdir`, `path.join`
- JSON: `json.parse`, `json.stringify`
- UI: `ui.element`, `ui.text`, `ui.renderToHtml`
- HTTP: `http.server().use/get/post/listen`

## 8) Complete folder and file structure
Provided directly in this repository.

## 9) Complete source code for all files
The repository contains the complete implementation.

## 10) Installation instructions
`go build -o airt ./cmd/airt`

## 11) Usage examples written in .aa
`examples/app/main.aa`, `examples/app/components.aa`

## 12) Error handling behavior
Parser and runtime errors are surfaced as `AirtError: ...` in CLI output.

## 13) Future extension points
Bytecode VM, type checker, package registry, deterministic scheduler, SSR hydration protocol.
