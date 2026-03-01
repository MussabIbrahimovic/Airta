# Airt Language (Native Runtime)

Airt is a standalone programming language with `.aa` source files, its own parser/evaluator, and a native CLI implemented in Go (no Node.js dependency).

## Install
```bash
go build -o airt ./cmd/airt
```

## CLI
```bash
./airt run examples/app/main.aa
./airt serve examples/app/main.aa
./airt build examples/app/main.aa dist
./airt validate examples/app/main.aa
./airt format examples/app/main.aa
```

See `docs/AIRT_SPEC.md` for the full language specification.
