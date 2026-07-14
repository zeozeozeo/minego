# Packet adapter status for Minecraft 26.2

This data directory is generated independently. Packet structs are deliberately not copied from another release: add them in a version-owned Go package after reviewing the generated packet IDs and wire schemas. The normalized handshake/status/login packet-ID map may be added to `versions`, but configuration and play adapters require golden encode/decode coverage before a pack is registered in `versions.Supported`.
