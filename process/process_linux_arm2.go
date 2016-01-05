// +build linux
// +build arm arm64

// FIXME(ssx): I added arm64 but not tested

package process

const (
	ClockTicks = 100  // C.sysconf(C._SC_CLK_TCK)
	PageSize   = 4096 // C.sysconf(C._SC_PAGE_SIZE)
)
