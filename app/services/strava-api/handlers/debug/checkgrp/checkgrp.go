// Package checkgrp maintains the group of handlers for health checking.
package checkgrp

import (
	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
}
