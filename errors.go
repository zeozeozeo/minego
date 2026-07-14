package minego

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zeozeozeo/minego/version"
)

var (
	ErrNotConnected       = errors.New("minego: not connected")
	ErrAlreadyConnected   = errors.New("minego: already connected")
	ErrInvalidSlot        = errors.New("minego: hotbar slot must be between 0 and 8")
	ErrUnreachable        = errors.New("minego: goal is unreachable")
	ErrSearchExhausted    = errors.New("minego: search exhausted")
	ErrNoPlacementItem    = errors.New("minego: no matching placeable item in the hotbar")
	ErrNoPlacementFace    = errors.New("minego: no solid neighboring face supports placement")
	ErrUnsupportedFeature = errors.New("minego: unsupported feature")
	ErrUnsupportedVersion = errors.New("minego: unsupported Minecraft version")
)

type UnsupportedFeatureError struct {
	Feature version.Feature
	Version version.Descriptor
}

func (e *UnsupportedFeatureError) Error() string {
	return fmt.Sprintf("%v: %s on %s", ErrUnsupportedFeature, e.Feature, e.Version.Name)
}
func (e *UnsupportedFeatureError) Unwrap() error { return ErrUnsupportedFeature }

type UnsupportedVersionError struct {
	Name      string
	Protocol  int32
	Supported []version.Descriptor
}

func (e *UnsupportedVersionError) Error() string {
	requested := e.Name
	if requested == "" {
		requested = fmt.Sprintf("protocol %d", e.Protocol)
	}
	names := make([]string, len(e.Supported))
	for i, descriptor := range e.Supported {
		names[i] = fmt.Sprintf("%s (%d)", descriptor.Name, descriptor.Protocol)
	}
	return fmt.Sprintf("%v: %s; supported: %s", ErrUnsupportedVersion, requested, strings.Join(names, ", "))
}
func (e *UnsupportedVersionError) Unwrap() error { return ErrUnsupportedVersion }
