package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/slavash/throttled"
)

// ****************************************************************************
// THIS IS STRAIGHTFORWARD IMPLEMENTATION TO DEMONSTRATE THE LIBRARY USAGE ONLY
// ****************************************************************************

const (
	exitCmd          = "exit"
	defaultConnLimit = (980933 - burst) / 3 // I want to download my file (size: 980933) in 3 sec
	burst            = 32 * 1024            // using default io.Copy buffer size as the allowed burst
	defaultTimeout   = 3 * time.Second
)

var connections []*throttled.LimitedConnection

func main() {
	endPoint := "localhost:7777"
	l, err := net.Listen("tcp4", endPoint)
	if err != nil {
		fmt.Println(err)
		return
	}

	limited := throttled.NewListener(l)
	limited.SetLimits(defaultConnLimit*3, defaultConnLimit)

	defer func() {
		if e := limited.Close(); e != nil {
			fmt.Printf("faield to shotdown the server: %s\n", e)
		}
	}()

	fmt.Printf("Listening on %s\n", l.Addr())

	for {
		c, err := limited.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		connections = append(connections, c)
		go handleConnection(c)
	}
}

func handleConnection(c *throttled.LimitedConnection) {

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer func() {
		if e := c.Close(); e != nil {
			fmt.Printf("faield to close connection: %s\n", e)
		}
		cancel()
	}()

	fmt.Printf("client connected from %s\n", c.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		cmd := strings.TrimSpace(netData)
		fmt.Printf("Received command: %s\n", cmd)

		if cmd == exitCmd {
			fmt.Printf("Closing connection for the client %s\n", c.RemoteAddr().String())
			break
		}

		// example of changing limits in runtime (applies to all existing connections)
		if len(cmd) > 5 && cmd[:4] == "setl" {
			limit, err := strconv.Atoi(cmd[5:])
			if err != nil {
				fmt.Printf("invalid limit value: %s\n", cmd[5:])
				_, _ = fmt.Fprintf(c, "invalid limit value: %s\n", cmd[5:])
				continue
			}
			if limit == 0 {
				limit = defaultConnLimit
			}

			c.SetLimit(limit)
			fmt.Printf("Rate limit changed to %d\n", limit)
		}

		if len(cmd) > 4 && cmd[:3] == "get" {
			fileName := cmd[4:]
			err := serveFile(ctx, c, fileName)
			if err != nil {
				fmt.Printf("failed to serve data: %s\n", err)
				_, _ = fmt.Fprintf(c, "failed to serve data: %s\n", err)
			}
		}
	}
}

func serveFile(_ context.Context, c *throttled.LimitedConnection, name string) error {

	var sent int64
	defer func(start time.Time) {
		fmt.Printf("Sent %d bytes in %.3fs\n", sent, time.Since(start).Seconds())
	}(time.Now())

	fd, err := os.Open(name)
	if err != nil {
		return err
	}

	sent, err = io.Copy(c, fd)

	if err != nil {
		return err
	}

	return nil
}
