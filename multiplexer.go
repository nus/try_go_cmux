package main

import (
	"bytes"
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"github.com/soheilhy/cmux"
)

func serveHTTP(l net.Listener) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "The HTTP handler is called.")
	})
	err := http.Serve(l, nil) // Always returns a non-nil error.
	if err != cmux.ErrListenerClosed {
		panic(fmt.Sprintf("Failed to http.Server. err: %s", err.Error()))
	}
}

func serveMyProto(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			if err != cmux.ErrListenerClosed {
				panic(fmt.Sprintf("Failed to l.Accept(). err: %s", err.Error()))
			}
		}
		go func() {
			defer conn.Close()

			MY_PROTO := []byte("myprotocol\r\n")
			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			buf := make([]byte, 1024)
			size, err := reader.Read(buf)
			if err != nil {
				panic(fmt.Sprintf("Failed to reader.Read(). err: %s", err.Error()))
			} else if !bytes.HasPrefix(buf, MY_PROTO) {
				panic(fmt.Sprintf("Unexpected protocol."))
			} else if _, err = writer.Write(buf[:size]); err != nil {
				panic(fmt.Sprintf("Failed to writer.Write(). err: %s", err.Error()))
			} else if err = writer.Flush(); err != nil {
				panic(fmt.Sprintf("Failed to writer.Flush(). err: %s", err.Error()))
			}
		}()
	}
}

func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(fmt.Sprintf("Failed to net.Listen(). err: %s", err.Error()))
	}

	m := cmux.New(l)

	my_proto := m.Match(cmux.PrefixMatcher([]string{"myprotocol\r\n",}...))
	httpl := m.Match(cmux.HTTP1Fast())

	go serveHTTP(httpl)
	go serveMyProto(my_proto)

	err = m.Serve()
	if !strings.Contains(err.Error(), "use of closed network connection") {
		panic(fmt.Sprintf("Failed to m.Server(). err: %s", err.Error()))
	}
}
