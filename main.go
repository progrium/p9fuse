package main

import (
	"flag"
	"io"
	"log"
	"net"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hugelgupf/p9/p9"
)

func main() {
	debug := flag.Bool("debug", false, "print debug data")

	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatal("Usage:\n  p9fs [-debug] 9PADDR MOUNTPOINT")
	}

	conn, err := net.Dial("tcp", flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	client, err := p9.NewClient(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	root, err := client.Attach("/")
	if err != nil {
		log.Fatal(err)
	}

	opts := &fs.Options{}
	opts.Debug = *debug
	if !*debug {
		log.SetOutput(io.Discard)
	}

	server, err := fs.Mount(flag.Arg(1), &node{file: root, path: ""}, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}

	server.Wait()
}
