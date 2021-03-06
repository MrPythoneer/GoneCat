package gonecat

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type TcpCat struct {
	*GCArguments
	Address *net.TCPAddr
}

func (tc *TcpCat) Execute() error {
	if tc.Listening {
		return tc.listen()
	}

	return tc.connect()
}

func (tc *TcpCat) listen() error {
	listener, err := net.ListenTCP(tc.Network, tc.Address)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go tc.handle(&GCCon{conn})
	}
}

func (tc *TcpCat) connect() error {
	conn, err := net.DialTCP(tc.Network, nil, tc.Address)
	if err != nil {
		return err
	}

	tc.handle(&GCCon{conn})
	return nil
}

func (tc *TcpCat) handle(conn *GCCon) {
	defer conn.Close()
	defer tc.Output.Close()

	if tc.ReadPipe {
		tc.streamPipe(conn)
	}

	if tc.ReadStdin {
		go tc.streamStdin(conn)
	}

	io.Copy(tc.Output, conn)
}

func (tc *TcpCat) streamPipe(conn *GCCon) {
	for {
		_, err := io.CopyN(conn, os.Stdin, int64(tc.BufferSize))
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
	}
}

func (tc *TcpCat) streamStdin(conn *GCCon) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		str := scanner.Text()

		if tc.SendCRLF {
			fmt.Fprintln(conn, str)
		} else {
			conn.Write([]byte(str))
		}
	}
}
