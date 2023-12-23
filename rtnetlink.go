package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/mdlayher/netlink"
)

type IfStats struct {
	RxPackets uint64
	TxPackets uint64
	RxBytes   uint64
	TxBytes   uint64
	// RxErrors   uint64
	// TxErrors   uint64
	// RxDropped  uint64
	// TxDropped  uint64
	// Multicast  uint64
	// Collisions uint64

	// /* detailed rx_errors: */
	// RxLengthErrors uint64
	// RxOverErrors   uint64
	// RxCrcErrors    uint64
	// RxFrameErrors  uint64
	// RxFifoErrors   uint64
	// RxMissedErrors uint64

	// /* detailed tx_errors */
	// TxAbortedErrors   uint64
	// TxCarrierErrors   uint64
	// TxFifoErrors      uint64
	// TxHeartbeatErrors uint64
	// TxWindowErrors    uint64

	// /* for cslip etc */
	// RxCompressed uint64
	// TxCompressed uint64
	// RxNohandler  uint64

	// RxOtherhostDropped uint64
}

type NetifStats struct {
	c   *netlink.Conn
	req netlink.Message
}

func NewNetifStats(name string) (*NetifStats, error) {
	netif, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	ns := &NetifStats{}

	// Dial rtnetlink socket.
	const NETLINK_ROUTE = 0x00
	ns.c, err = netlink.Dial(NETLINK_ROUTE, nil)
	if err != nil {
		return nil, err
	}

	// Prepare rtnetlink message.
	ifstatsMsg := struct {
		Family     uint8
		pad1       uint8
		pad2       uint16
		IfIndex    uint32
		FilterMask uint32
	}{
		Family:     0, // AF_UNSPEC
		IfIndex:    uint32(netif.Index),
		FilterMask: (1 << 0), // IFLA_STATS_LINK_64
	}

	data := &bytes.Buffer{}
	binary.Write(data, binary.LittleEndian, &ifstatsMsg)

	const RTM_GETSTATS = 0x5e
	ns.req.Header.Type = RTM_GETSTATS
	ns.req.Header.Flags = netlink.Request | netlink.DumpFiltered
	ns.req.Data = data.Bytes()

	return ns, nil
}

func (ns *NetifStats) Read(p *IfStats) error {
	msgs, err := ns.c.Execute(ns.req)
	if err != nil {
		return err
	}

	if len(msgs) < 1 {
		return fmt.Errorf("empty response")
	}

	b := bytes.NewBuffer(msgs[0].Data[16:])
	binary.Read(b, binary.LittleEndian, p)

	return nil
}

func (ns *NetifStats) Close() error {
	ns.c.Close()
	return nil
}

// type IfStats struct {
// 	RxPackets uint64
// 	TxPackets uint64
// 	RxBytes   uint64
// 	TxBytes   uint64
// 	// RxErrors   uint64
// 	// TxErrors   uint64
// 	// RxDropped  uint64
// 	// TxDropped  uint64
// 	// Multicast  uint64
// 	// Collisions uint64

// 	// /* detailed rx_errors: */
// 	// RxLengthErrors uint64
// 	// RxOverErrors   uint64
// 	// RxCrcErrors    uint64
// 	// RxFrameErrors  uint64
// 	// RxFifoErrors   uint64
// 	// RxMissedErrors uint64

// 	// /* detailed tx_errors */
// 	// TxAbortedErrors   uint64
// 	// TxCarrierErrors   uint64
// 	// TxFifoErrors      uint64
// 	// TxHeartbeatErrors uint64
// 	// TxWindowErrors    uint64

// 	// /* for cslip etc */
// 	// RxCompressed uint64
// 	// TxCompressed uint64
// 	// RxNohandler  uint64

// 	// RxOtherhostDropped uint64
// }
