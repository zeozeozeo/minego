# Generated data implementation

Do not invoke this package directly. Run `go generate ./...` at the repository
root. `tools/mcgen` obtains and verifies the official artifacts, creates the raw
inputs in this directory, runs this Go generator, and then regenerates packet
methods. Raw inputs are intentionally ignored; generated Go sources are tracked.
