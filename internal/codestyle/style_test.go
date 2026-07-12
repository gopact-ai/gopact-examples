package codestyle

import (
	"testing"

	"github.com/gopact-ai/gopact/gopacttest"
)

// TestCodeStyle verifies that every public example follows the shared repository style contract.
func TestCodeStyle(t *testing.T) {
	gopacttest.RequireCodeStyle(t, "../..")
}
