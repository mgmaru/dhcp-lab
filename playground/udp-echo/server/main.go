package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	// convert to Address
	addr, err := net.ResolveUDPAddr("udp4", ":9067")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// create socket and bind
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	// display socket Address
	socketAddr := conn.LocalAddr()
	fmt.Println(socketAddr)

	// read message from UDPclient
	buf := make([]byte, 1000) // length 1000byte
	for {
		n, clientAddr, err := conn.ReadFromUDPAddrPort(buf)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		fmt.Println(clientAddr, string(buf[:n]))
	}
}
