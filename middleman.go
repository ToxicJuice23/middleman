package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

func handleClient(client net.Conn, host net.Conn) {
	returned := false
	go func() {
		for {
			buf := make([]byte, 1024)
			_, err := client.Read(buf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error on recv().\n%s\n", err.Error())
				client.Close()
				returned = true
				return
			}
			host.Write(buf)
		}
	}()

	go func() {
		for {
			buf := make([]byte, 2048)
			_, err := host.Read(buf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error on read().\n%s\n", err.Error())
				client.Close()
				host.Close()
				returned = true
				return
			}
			_, err = client.Write(buf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error on write().\n%s\n", err.Error())
				client.Close()
				returned = true
				return
			}
		}
	}()
	for {
		if returned {
			return
		}
	}
}

func main() {
	var host net.Conn
	host = nil

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	} else {
		port = ":" + port
	}

	serverConn, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occured when listening for connections: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Started middleman server on port %s\nWaiting for the host to connect...\n", port)

serverInit:
	host, err = serverConn.Accept()
	fmt.Fprintf(os.Stdout, "recieved connection from: %s\n", host.RemoteAddr().String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error occured when accepting a connection: %s\n", err.Error())
		goto serverInit
	}
	buf := make([]byte, 100)
	wanted := []byte("host")
	host.Read(buf)
	if !bytes.Contains(buf, wanted) {
		fmt.Fprintf(os.Stderr, "Recieved a client instead of server.\n")
		host.Close()
		host = nil
		goto serverInit
	}
	host.Write([]byte("Ok"))

	for {
		client, err := serverConn.Accept()

		fmt.Fprintf(os.Stdout, "recieved connection from: %s\n", client.RemoteAddr().String())
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occured when accepting a connection: %s\n", err.Error())
			continue
		}

		buf = make([]byte, 100)
		wanted = []byte("client")
		client.Read(buf)
		if !bytes.Contains(buf, wanted) {
			fmt.Fprintf(os.Stderr, "Error: was expecting new client\n")
			client.Close()
			continue
		}
		client.Write([]byte("Ok"))
		handleClient(client, host)
	}
}
