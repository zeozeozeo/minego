package blocks_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/zeozeozeo/minego/internal/data/blocks"
)

type blockJSON struct {
	Properties map[string][]string `json:"properties"`
	States     []stateJSON         `json:"states"`
}

type stateJSON struct {
	ID         int32             `json:"id"`
	Default    bool              `json:"default,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

func loadBlocksJSON(t *testing.T) map[string]blockJSON {
	data, err := os.ReadFile("../generate/blocks.json")
	if err != nil {
		t.Skipf("raw generator input is not present: %v", err)
	}
	var result map[string]blockJSON
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse blocks.json: %v", err)
	}
	return result
}

func TestAllBlockStates(t *testing.T) {
	blocksData := loadBlocksJSON(t)

	totalStates := 0
	failedStates := 0

	for blockName, blockData := range blocksData {
		blockID := blocks.BlockID(blockName)
		if blockID == -1 {
			t.Errorf("block %q not found in registry", blockName)
			continue
		}

		for _, state := range blockData.States {
			totalStates++
			calculated := blocks.StateID(int(blockID), state.Properties)
			if calculated != state.ID {
				failedStates++
				if failedStates <= 10 {
					t.Errorf("%s: expected state %d, got %d for props %v",
						blockName, state.ID, calculated, state.Properties)
				}
			}
		}
	}

	if failedStates > 10 {
		t.Errorf("... and %d more failures (total: %d/%d states failed)",
			failedStates-10, failedStates, totalStates)
	}

	t.Logf("tested %d block states", totalStates)
}

func TestDefaultStates(t *testing.T) {
	blocksData := loadBlocksJSON(t)

	for blockName, blockData := range blocksData {
		blockID := blocks.BlockID(blockName)
		if blockID == -1 {
			continue
		}

		// find the default state in JSON
		var expectedDefaultID int32 = -1
		for _, state := range blockData.States {
			if state.Default {
				expectedDefaultID = state.ID
				break
			}
		}

		// blocks without explicit default use first state
		if expectedDefaultID == -1 && len(blockData.States) > 0 {
			expectedDefaultID = blockData.States[0].ID
		}

		actualDefaultID := blocks.DefaultStateID(blockID)
		if actualDefaultID != expectedDefaultID {
			t.Errorf("%s: expected default state %d, got %d",
				blockName, expectedDefaultID, actualDefaultID)
		}
	}
}

func TestStateProperties(t *testing.T) {
	blocksData := loadBlocksJSON(t)

	// test a sample of states
	testedCount := 0
	for blockName, blockData := range blocksData {
		if len(blockData.Properties) == 0 {
			continue
		}

		blockID := blocks.BlockID(blockName)
		if blockID == -1 {
			continue
		}

		// test first and last state of each block with properties
		for _, idx := range []int{0, len(blockData.States) - 1} {
			if idx >= len(blockData.States) {
				continue
			}
			state := blockData.States[idx]

			gotBlockID, gotProps := blocks.StateProperties(int(state.ID))
			if gotBlockID != blockID {
				t.Errorf("StateProperties(%d): expected block %d (%s), got %d",
					state.ID, blockID, blockName, gotBlockID)
				continue
			}

			for prop, expectedValue := range state.Properties {
				if gotProps[prop] != expectedValue {
					t.Errorf("StateProperties(%d): %s expected %q=%q, got %q",
						state.ID, blockName, prop, expectedValue, gotProps[prop])
				}
			}
			testedCount++
		}

		// limit to avoid extremely long tests
		if testedCount > 2000 {
			break
		}
	}

	t.Logf("tested %d state reverse lookups", testedCount)
}
