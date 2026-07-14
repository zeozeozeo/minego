# Migration to the context-aware API

MineGo's version selection is now a release string rather than a generated data
pack:

```go
// Automatic status detection (recommended).
bot, err := minego.New(minego.Config{Address: address})

// Or pin a release without importing versions or generated data.
bot, err := minego.New(minego.Config{Address: address, Version: "26.2"})
```

After automatic detection, use `bot.Version()` for immutable metadata and
`bot.Supports(minego.FeatureNavigation)` for capability checks; `bot.Require`
returns a typed unsupported-feature error. An unsupported
pinned name fails in `New`; an unsupported detected protocol fails in
`Connect`. Both can be recognized with `errors.Is(err,
minego.ErrUnsupportedVersion)` and inspected as `*minego.UnsupportedVersionError`.

Code that used `bot.Version().StateID` should use version-neutral world and
service operations. `bot.HasBlock(name)` is available for simple validation.
Long-running navigation, mining, placement, chat, and inventory calls take a
`context.Context`; cancel it to stop the action and release its controls.
