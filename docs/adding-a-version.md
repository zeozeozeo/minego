# Adding a Minecraft version

1. Run `go run ./tools/mcgen -version <version>` to obtain verified official
   reports, collision dumps, translations, tags, and regenerated tables.
2. Keep the generated data isolated from existing packs. Add a value in
   `versions/` implementing `version.Pack`; its `Protocol` method and packet
   registry must match the new `packets.json` report.
3. Diff the official packet report against the prior version. Generated IDs are
   mechanical, but changed wire fields require a new or adapted packet struct.
   Never reuse a prior codec merely because its packet name is unchanged.
4. Add golden codec tests for every changed schema and update connection-state
   handlers only behind version feature checks. Do not branch public bot APIs on
   a protocol number.
5. Run `go test -race ./...`, `go vet ./...`, and the gated offline dedicated-
   server integration test for the new version before exposing the pack.

Raw generator inputs and downloaded jars are ignored; generated Go tables and
the version pack are committed. This makes review show exactly which registry,
state, collision, and packet mappings changed.
