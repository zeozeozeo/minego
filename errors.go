package minego

import "errors"

var (
	ErrNotConnected     = errors.New("minego: not connected")
	ErrAlreadyConnected = errors.New("minego: already connected")
	ErrInvalidSlot      = errors.New("minego: hotbar slot must be between 0 and 8")
	ErrUnreachable      = errors.New("minego: goal is unreachable")
	ErrSearchExhausted  = errors.New("minego: search exhausted")
	ErrNoPlacementItem  = errors.New("minego: no matching placeable item in the hotbar")
	ErrNoPlacementFace  = errors.New("minego: no solid neighboring face supports placement")
)
