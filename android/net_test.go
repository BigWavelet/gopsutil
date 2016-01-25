package android

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNetworkStats(t *testing.T) {
	PROC_NET_STATS = "testdata/qtaguid_stats"
	Convey("Get network stats", t, func() {
		nss, err := NetworkStats()
		So(err, ShouldBeNil)

		var rxBytes, txBytes uint64 = 0, 0
		for _, ns := range nss {
			if ns.Uid == 1013 {
				rxBytes += ns.RecvBytes
				txBytes += ns.SendBytes
			}
		}
		So(rxBytes, ShouldEqual, 0)
		So(txBytes, ShouldEqual, 1716)
	})
}
