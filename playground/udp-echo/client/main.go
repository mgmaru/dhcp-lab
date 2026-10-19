package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"time"
)

type networkArgs struct {
	serverHost string
	serverPort string
	kinds      int
	id         uint
	message    string
}

func setNetworkArgs(method string) networkArgs {
	if method == "0" {
		serverHost := flag.String("addr", "127.0.0.1", "Set the server address. The default is 127.0.0.1")
		serverPort := flag.String("port", "9067", "Set the server port. The default is 9067.")
		kinds := flag.Int("kinds", 1, "kinds flag 1(message) or 2(ping) default 1") // 1byte -> 0-255
		id := flag.Uint("ID", 0, "ID uint32 default 0")                             // uint64 but the client requires 32bit(uint32)
		message := flag.String("message", "", "input send message")

		flag.Parse()

		return networkArgs{
			serverHost: *serverHost,
			serverPort: *serverPort,
			kinds:      *kinds,
			id:         *id,
			message:    *message,
		}

	} else { // broadcast
		serverPort := flag.String("port", "9067", "Set the server port. The default is 9067.")
		kinds := flag.Int("kinds", 1, "kinds flag 1(message) or 2(ping) default 1") // 1byte -> 0-255
		id := flag.Uint("ID", 0, "ID uint32 default 0")                             // uint64 but the client requires 32bit(uint32)
		message := flag.String("message", "", "input send message")

		flag.Parse()

		return networkArgs{
			serverHost: "255.255.255.255",
			serverPort: *serverPort,
			kinds:      *kinds,
			id:         *id,
			message:    *message,
		}
	}
}

func validateNetworkArgs(args networkArgs) {
	if net.IP(args.serverHost) == nil {
		fmt.Fprintln(os.Stderr, "The network address is invalid.")
		os.Exit(1)
	}
	if args.kinds != 1 && args.kinds != 2 {
		fmt.Fprintln(os.Stderr, "This is not a valid type. Valid types are 1（message） or 2（ping）.")
		os.Exit(1)
	}
	if args.id > math.MaxUint32 {
		fmt.Fprintln(os.Stderr, "This is not a valid type. The valid ID range is from 0 to the maximum value of uint32.")
		os.Exit(1)
	}
}

func main() {
	// Args
	// method setting
	fmt.Println("Please select either 0 (unicast) or 1 (broadcast).：", os.Args)
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "The number of arguments is incorrect. There is only one argument.")
		os.Exit(1)
	}
	if os.Args[0] != "0" && os.Args[0] != "1" {
		fmt.Fprintln(os.Stderr, "Invalid argument. The argument must be 0 (unicast) or 1 (broadcast).")
		os.Exit(1)
	}

	method := os.Args[0]

	args := setNetworkArgs(method)

	// Argument validation
	validateNetworkArgs(args)

	// byte length
	kindsBytelength := 1
	idByteLength := 4
	messageByteLength := len(args.message)
	totalByteLength := kindsBytelength + idByteLength + messageByteLength

	// byte Length validation
	if totalByteLength > 1000 {
		fmt.Fprintf(os.Stderr, "Invalid byte length. Valid byte sequence length is 1000 or less.")
		os.Exit(1)
	}

	// convert to client Address
	clientAddr, err := net.ResolveUDPAddr("udp4", ":9068")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// convert to server Address
	serverAddr, err := net.ResolveUDPAddr("udp4", args.serverHost+":"+args.serverPort)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// create socket and bind
	conn, err := net.ListenUDP("udp4", clientAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// write the message to the buffer
	buf := make([]byte, totalByteLength)
	buf[0] = byte(args.kinds)
	binary.BigEndian.PutUint32(buf[1:5], uint32(args.id))
	copy(buf[5:], []byte(args.message))

	// send packet to server
	_, err = conn.WriteToUDP(buf, serverAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// display send message to UDP server
	fmt.Println("kinds：" + string(buf[0]))
	fmt.Println("ID：" + string(buf[1:5]))
	fmt.Println("message：" + args.message)

	// read message from UDP server
	readBuf := make([]byte, 1000)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second)) // Set the timeout period to 3 seconds.

	n, server, err := conn.ReadFromUDPAddrPort(readBuf)
	// timout Error
	if errors.Is(err, os.ErrDeadlineExceeded) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else { // other Error
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// display message from UDP server
	fmt.Println(server, string(buf[:n]))
}
