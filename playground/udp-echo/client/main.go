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

type ReceivedData struct {
	kinds   int
	id      uint
	message string
}

func setSubCmds(method string, args []string) networkArgs {

	subCmds := flag.NewFlagSet("sub", flag.ExitOnError)
	if method == "0" {
		serverHost := subCmds.String("addr", "127.0.0.1", "Set the server address. The default is 127.0.0.1")
		serverPort := subCmds.String("port", "9067", "Set the server port. The default is 9067.")
		kinds := subCmds.Int("kinds", 1, "kinds flag 1(message) or 2(ping) default 1") // 1byte -> 0-255
		id := subCmds.Uint("id", 0, "ID uint32 default 0")                             // uint64 but the client requires 32bit(uint32)
		message := subCmds.String("message", "", "input send message")

		subCmds.Parse(args)

		return networkArgs{
			serverHost: *serverHost,
			serverPort: *serverPort,
			kinds:      *kinds,
			id:         *id,
			message:    *message,
		}

	} else { // broadcast
		serverPort := subCmds.String("port", "9067", "Set the server port. The default is 9067.")
		kinds := subCmds.Int("kinds", 1, "kinds flag 1(message) or 2(ping) default 1") // 1byte -> 0-255
		id := subCmds.Uint("id", 0, "ID uint32 default 0")                             // uint64 but the client requires 32bit(uint32)
		message := subCmds.String("message", "", "input send message")

		subCmds.Parse(args)

		return networkArgs{
			serverHost: "255.255.255.255",
			serverPort: *serverPort,
			kinds:      *kinds,
			id:         *id,
			message:    *message,
		}
	}
}

// validate method
func validateMethod(method string) string {
	if method != "0" && method != "1" {
		fmt.Fprintln(os.Stderr, "Invalid argument. The argument must be 0 (unicast) or 1 (broadcast).")
		os.Exit(1)
	}
	validMethod := method
	return validMethod
}

func validateSubCmds(args networkArgs) {
	if net.ParseIP(args.serverHost) == nil {
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

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "The argument is missing.")
		os.Exit(1)
	}

	validMethod := validateMethod(os.Args[1])

	// set subCmds
	subCmds := setSubCmds(validMethod, os.Args[2:])
	validateSubCmds(subCmds)

	// byte length
	kindsBytelength := 1
	idByteLength := 4
	messageByteLength := len(subCmds.message)
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
	serverAddr, err := net.ResolveUDPAddr("udp4", subCmds.serverHost+":"+subCmds.serverPort)
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

	defer conn.Close()

	// encode and write the message to the buffer
	buf := make([]byte, totalByteLength)
	buf[0] = byte(subCmds.kinds)
	binary.BigEndian.PutUint32(buf[1:5], uint32(subCmds.id))
	copy(buf[5:], []byte(subCmds.message))

	// send packet to server
	_, err = conn.WriteToUDP(buf, serverAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// display send message to UDP server
	fmt.Printf("kinds：%d\n", subCmds.kinds)
	fmt.Printf("ID：%d\n", subCmds.id)
	fmt.Printf("message：%s\n", subCmds.message)

	// read message from UDP server
	readBuf := make([]byte, 1000)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second)) // Set the timeout period to 3 seconds.

	readBufLen, server, err := conn.ReadFromUDPAddrPort(readBuf)
	if errors.Is(err, os.ErrDeadlineExceeded) { // timout Error
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else if err != nil { // other Error
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Check the length of the received buffer
	if readBufLen < 5 { // Not long enough
		fmt.Fprintln(os.Stderr, "The received message is too short. The minimum length required is 5 bytes.")
		os.Exit(1)
	}

	// decode
	receivedKinds := int(readBuf[0])
	receivedId := uint(binary.BigEndian.Uint32(readBuf[1:5]))
	receivedMessage := string(readBuf[5:readBufLen])

	// validation
	if receivedKinds != 1 && receivedKinds != 2 {
		fmt.Fprintln(os.Stderr, "This is an invalid type. The valid types are 1 or 2.")
		os.Exit(1)
	}
	if receivedKinds != subCmds.kinds {
		fmt.Fprintln(os.Stderr, "The type of message sent and the type of message received are different.")
		os.Exit(1)
	}
	if receivedId != subCmds.id {
		fmt.Fprintln(os.Stderr, "The ID sent and the ID received are different.")
		os.Exit(1)
	}
	if receivedMessage != subCmds.message {
		fmt.Fprintln(os.Stderr, "The message I sent and the message I received are different.")
		os.Exit(1)
	}

	receivedData := ReceivedData{
		kinds:   receivedKinds,
		id:      receivedId,
		message: receivedMessage,
	}

	// display message from UDP server
	fmt.Printf("server：%s\n", server)
	fmt.Printf("kinds：%d\n", receivedData.kinds)
	fmt.Printf("ID：%d\n", receivedData.id)
	fmt.Printf("message：%s\n", receivedData.message)
}
