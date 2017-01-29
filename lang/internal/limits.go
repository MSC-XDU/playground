package lang

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
)

func DefaultResourceLimits() container.Resources {
	cputime, _ := units.ParseUlimit("cpu=10:10")
	return container.Resources{
		CPUPeriod:  100000,
		CPUQuota:   30000,
		Memory:     30 * units.MiB,
		MemorySwap: 100 * units.MiB,
		Ulimits:    []*units.Ulimit{cputime},
	}
}
