package android

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	PROC_NET_STATS = "/proc/net/xt_qtaguid/stats"
)

type NetStat struct {
	Index            int
	Interface        string
	SocketTag        string
	Uid              int
	IsBackground     bool
	RecvBytes        uint64
	RecvPackets      uint64
	SendBytes        uint64
	SendPackets      uint64
	RecvTcpBytes     uint64
	RecvTcpPackets   uint64
	RecvUdpBytes     uint64
	RecvUdpPackets   uint64
	RecvOtherBytes   uint64
	RecvOtherPackets uint64
	SendTcpBytes     uint64
	SendTcpPackets   uint64
	SendUdpBytes     uint64
	SendUdpPackets   uint64
	SendOtherBytes   uint64
	SendOtherPackets uint64
}

// http://stackoverflow.com/questions/15163549/interpreting-android-xt-qtaguid-stats
// here RecvBytes include tcp header
func NetworkStats() (nss []NetStat, err error) {
	fd, err := os.Open(PROC_NET_STATS)
	if err != nil {
		return
	}
	defer fd.Close()
	rd := bufio.NewReader(fd)
	nss = make([]NetStat, 0, 50)

	rd.ReadLine() // ignore the first line
	for {
		line, _, er := rd.ReadLine()
		if er != nil {
			if er == io.EOF {
				return nss, nil
			}
			return nil, er
		}
		ns := NetStat{}
		_, err = fmt.Sscanf(string(line), "%d %s %s %d %t"+strings.Repeat(" %d", 16),
			&ns.Index, &ns.Interface, &ns.SocketTag, &ns.Uid, &ns.IsBackground,
			&ns.RecvBytes, &ns.RecvPackets, &ns.SendBytes, &ns.SendPackets,
			&ns.RecvTcpBytes, &ns.RecvTcpPackets,
			&ns.RecvUdpBytes, &ns.RecvUdpPackets,
			&ns.RecvOtherBytes, &ns.RecvOtherPackets,
			&ns.SendTcpBytes, &ns.SendTcpPackets,
			&ns.SendUdpBytes, &ns.SendUdpPackets,
			&ns.SendOtherBytes, &ns.SendOtherPackets)
		if err != nil {
			return
		}
		nss = append(nss, ns)
	}
}

// Not working well in some devices
// traffix for android: http://keepcleargas.bitbucket.org/2013/10/12/android-App-Traffic.html
/*
	tcpRecv := fmt.Sprintf("/proc/uid_stat/%d/tcp_rcv", uid)
	tcpSend := fmt.Sprintf("/proc/uid_stat/%d/tcp_snd", uid)
	rcv, err = readUint64FromFile(tcpRecv)
	if err != nil {
		return
	}
	snd, err = readUint64FromFile(tcpSend)
	return
*/
