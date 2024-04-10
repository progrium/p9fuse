package main

import (
	"context"
	"log"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hugelgupf/p9/p9"
)

type node struct {
	fs.Inode
	file p9.File
	path string
}

var _ = (fs.NodeGetattrer)((*node)(nil))

func (r *node) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.Println("getattr", r.path)

	_, _, attrs, err := r.file.GetAttr(p9.AttrMaskAll)
	if err != nil {
		return sysErrno(err)
	}
	out.FromStat(sysStat(attrs))

	return 0
}

var _ = (fs.NodeReaddirer)((*node)(nil))

func (r *node) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	log.Println("readdir", r.path)

	_, f, err := r.file.Walk([]string{})
	if err != nil {
		return nil, sysErrno(err)
	}

	_, _, err = f.Open(p9.ReadOnly)
	if err != nil {
		return nil, sysErrno(err)
	}

	entries, err := f.Readdir(0, 1024) // todo: more than 1024...
	if err != nil {
		return nil, sysErrno(err)
	}

	if err := f.Close(); err != nil {
		return nil, sysErrno(err)
	}

	var fentries []fuse.DirEntry
	for _, entry := range entries {
		fentries = append(fentries, fuse.DirEntry{
			Name: entry.Name,
			Mode: fuseMode(entry.Type),
			Ino:  entry.QID.Path,
		})
	}

	return fs.NewListDirStream(fentries), 0
}

var _ = (fs.NodeOpendirer)((*node)(nil))

func (r *node) Opendir(ctx context.Context) syscall.Errno {
	log.Println("opendir", r.path)
	return 0
}

var _ = (fs.NodeLookuper)((*node)(nil))

func (r *node) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Println("lookup", r.path, name)

	qids, f, err := r.file.Walk([]string{name})
	if err != nil {
		return nil, sysErrno(err)
	}

	_, _, attrs, err := f.GetAttr(p9.AttrMaskAll)
	if err != nil {
		return nil, sysErrno(err)
	}
	out.FromStat(sysStat(attrs))

	return r.Inode.NewPersistentInode(ctx, &node{
		file: f,
		path: filepath.Join(r.path, name),
	}, fs.StableAttr{
		Mode: fuseMode(qids[0].Type),
		Ino:  qids[0].Path,
	}), 0
}

var _ = (fs.NodeOpener)((*node)(nil))

func (r *node) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	log.Println("open", r.path)

	_, f, err := r.file.Walk([]string{})
	if err != nil {
		return nil, 0, sysErrno(err)
	}

	_, fd, err := f.Open(openFlags(flags))
	if err != nil {
		return nil, 0, sysErrno(err)
	}

	return &handle{file: f, path: r.path, fd: fd}, 0, 0
}

//NodeSetattrer
//NodeWriter
//NodeFsyncer
//NodeMkdirer
//NodeCreater
//NodeUnlinker
//NodeRmdirer
//NodeRenamer
