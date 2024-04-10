package main

import (
	"context"
	"log"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hugelgupf/p9/p9"
)

type handle struct {
	file p9.File
	path string
	fd   uint32
}

var _ = (fs.FileReader)((*handle)(nil))

func (r *handle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.Println("read", r.path, r.fd)

	n, err := r.file.ReadAt(dest, off)
	if err != nil {
		return nil, sysErrno(err)
	}

	return fuse.ReadResultData(dest[:n]), 0
}

var _ = (fs.FileFlusher)((*handle)(nil))

func (r *handle) Flush(ctx context.Context) syscall.Errno {
	log.Println("flush", r.path, r.fd)

	if err := r.file.Close(); err != nil {
		return sysErrno(err)
	}

	return 0
}
