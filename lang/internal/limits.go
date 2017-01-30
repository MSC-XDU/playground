package lang

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-units"
)

var (
	cpuTime    int64 = 10
	cpuPeriod  int64 = 1000000
	cpuQuota   int64 = 300000
	memory     int64 = 30 * units.MiB
	memorySwap int64 = 100 * units.MiB
)

func init() {
	options := []struct {
		v *int64
		n string
	}{
		{&cpuTime, "L_CPU_TIME"}, {&cpuPeriod, "L_CPU_PERIOD"},
		{&cpuQuota, "L_CPU_QUOTA"}, {&memory, "L_MEM"},
		{&memorySwap, "L_TOTAL_MEM"},
	}

	for _, v := range options {
		err := getEnvInt(v.v, v.n)
		if err != nil {
			log.Panic(err)
		}
	}
}

func DefaultResourceLimits() container.Resources {
	cputime, _ := units.ParseUlimit(fmt.Sprintf("cpu=%d:%d", cpuTime, cpuTime))
	return container.Resources{
		CPUPeriod:  cpuPeriod,
		CPUQuota:   cpuQuota,
		Memory:     memory,
		MemorySwap: memorySwap,
		Ulimits:    []*units.Ulimit{cputime},
	}
}

func getEnvInt(dst *int64, name string) (err error) {
	if v := os.Getenv(name); v != "" {
		var n int
		n, err = strconv.Atoi(v)
		*dst = int64(n)
	}
	return
}
