package storage

import (
	"go.uber.org/dig"

	"github.com/ikeikeikeike/gocore/util/logger"
)

// Inject injects dependencies
func Inject(di *dig.Container) {
	// Injects
	var deps = []interface{}{
		newStorage,
	}

	for _, dep := range deps {
		if err := di.Provide(dep); err != nil {
			logger.Panicf("failed to process go-core storage injection: %s", err)
		}
	}
}
