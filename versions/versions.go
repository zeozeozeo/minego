package versions

import "github.com/zeozeozeo/minego/version"

var supported = []version.Pack{V1_21_11, V26_1, V26_2}

// Supported returns all compiled-in packs. The returned slice is independent
// and may be modified by the caller.
func Supported() []version.Pack { return append([]version.Pack(nil), supported...) }

func ByName(name string) (version.Pack, bool) {
	for _, pack := range supported {
		if pack.Name() == name {
			return pack, true
		}
	}
	return nil, false
}

func ByProtocol(protocol int32) (version.Pack, bool) {
	for _, pack := range supported {
		if pack.Protocol() == protocol {
			return pack, true
		}
	}
	return nil, false
}

func Descriptors() []version.Descriptor {
	result := make([]version.Descriptor, 0, len(supported))
	for _, pack := range supported {
		result = append(result, pack.Descriptor())
	}
	return result
}
