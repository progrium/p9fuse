package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hugelgupf/p9/linux"
	"github.com/hugelgupf/p9/p9"
)

type RemoteRoot struct {
	*RemoteNode
}

type RemoteNode struct {
	fs.Inode
	file p9.File
	path string
}

var _ = (fs.NodeGetattrer)((*RemoteNode)(nil))
var _ = (fs.NodeOpendirer)((*RemoteNode)(nil))
var _ = (fs.NodeReaddirer)((*RemoteNode)(nil))
var _ = (fs.NodeLookuper)((*RemoteNode)(nil))
var _ = (fs.NodeOpener)((*RemoteNode)(nil))

func (r *RemoteNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.Println("getattr", r.path)
	_, _, attrs, err := r.file.GetAttr(p9.AttrMaskAll)
	if err != nil {
		return ToErrno(err)
	}
	out.FromStat(ToStat(attrs))
	return 0
}

func (r *RemoteNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	log.Println("readdir", r.path)

	_, f, err := r.file.Walk([]string{})
	if err != nil {
		return nil, ToErrno(err)
	}

	_, _, err = f.Open(p9.ReadOnly)
	if err != nil {
		return nil, ToErrno(err)
	}

	entries, err := f.Readdir(0, 1024) // todo: more than 1024...
	if err != nil {
		return nil, ToErrno(err)
	}

	if err := f.Close(); err != nil {
		return nil, ToErrno(err)
	}

	var fentries []fuse.DirEntry
	for _, entry := range entries {
		fentries = append(fentries, fuse.DirEntry{
			Name: entry.Name,
			Mode: ToMode(entry.Type),
			Ino:  entry.QID.Path,
		})
	}

	return fs.NewListDirStream(fentries), 0
}

func (r *RemoteNode) Opendir(ctx context.Context) syscall.Errno {
	log.Println("opendir", r.path)
	return 0
}

func (r *RemoteNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Println("lookup", r.path, name)

	qids, f, err := r.file.Walk([]string{name})
	if err != nil {
		return nil, ToErrno(err)
	}

	_, _, attrs, err := f.GetAttr(p9.AttrMaskAll)
	if err != nil {
		return nil, ToErrno(err)
	}
	out.FromStat(ToStat(attrs))

	return r.Inode.NewPersistentInode(ctx, &RemoteNode{
		file: f,
		path: filepath.Join(r.path, name),
	}, fs.StableAttr{
		Mode: ToMode(qids[0].Type),
		Ino:  qids[0].Path,
	}), 0
}

func (r *RemoteNode) Open(ctx context.Context, flags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	log.Println("open", r.path)

	_, f, err := r.file.Walk([]string{})
	if err != nil {
		return nil, 0, ToErrno(err)
	}

	_, fd, err := f.Open(ToOpenFlags(flags))
	if err != nil {
		return nil, 0, ToErrno(err)
	}

	return &RemoteHandle{file: f, path: r.path, fd: fd}, 0, 0
}

type RemoteHandle struct {
	file p9.File
	path string
	fd   uint32
}

var _ = (fs.FileReader)((*RemoteHandle)(nil))
var _ = (fs.FileFlusher)((*RemoteHandle)(nil))

func (r *RemoteHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.Println("read", r.path, r.fd)

	n, err := r.file.ReadAt(dest, off)
	if err != nil {
		return nil, ToErrno(err)
	}

	return fuse.ReadResultData(dest[:n]), 0
}

func (r *RemoteHandle) Flush(ctx context.Context) syscall.Errno {
	log.Println("flush", r.path, r.fd)
	if err := r.file.Close(); err != nil {
		return ToErrno(err)
	}
	return 0
}

func main() {

	debug := flag.Bool("debug", false, "print debug data")
	flag.Parse()
	if len(flag.Args()) < 2 {
		log.Fatal("Usage:\n  p9fs MOUNTPOINT 9PADDR")
	}
	opts := &fs.Options{}
	opts.Debug = *debug

	conn, err := net.Dial("tcp", flag.Arg(1))
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

	// _, _, err = root.Open(p9.ReadOnly)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	server, err := fs.Mount(flag.Arg(0), &RemoteRoot{
		RemoteNode: &RemoteNode{file: root, path: ""},
	}, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	server.Wait()
}

func ToMode(t p9.QIDType) uint32 {
	switch t {
	case p9.TypeDir:
		return fuse.S_IFDIR
	case p9.TypeRegular:
		return fuse.S_IFREG
	case p9.TypeLink, p9.TypeSymlink:
		return fuse.S_IFLNK
	default:
		log.Panicf("unknown QID<->FUSE type: %v", t)
		return 0
	}
}

func ToOpenFlags(flags uint32) p9.OpenFlags {
	accessMode := flags & syscall.O_ACCMODE
	switch accessMode {
	case syscall.O_RDONLY:
		return p9.ReadOnly
	case syscall.O_WRONLY:
		return p9.WriteOnly
	case syscall.O_RDWR:
		return p9.ReadWrite
	default:
		log.Panicf("unknown access mode to open flag type: %v", accessMode)
		return 0
	}
}

func ToErrno(err error) syscall.Errno {
	log.Println("err:", err)
	switch err {
	case nil:
		return syscall.Errno(0)
	case os.ErrPermission:
		return syscall.EPERM
	case os.ErrExist:
		return syscall.EEXIST
	case os.ErrNotExist:
		return syscall.ENOENT
	case os.ErrInvalid:
		return syscall.EINVAL
	}

	switch t := err.(type) {
	case linux.Errno:
		return syscall.Errno(t)
	case syscall.Errno:
		return t
	case *os.SyscallError:
		return t.Err.(syscall.Errno)
	case *os.PathError:
		return ToErrno(t.Err)
	case *os.LinkError:
		return ToErrno(t.Err)
	}
	log.Println("can't convert error type:", err)
	return syscall.EINVAL
}

func ToStat(attr p9.Attr) *syscall.Stat_t {
	st := &syscall.Stat_t{}
	st.Mode = uint16(attr.Mode)
	st.Dev = int32(attr.RDev)
	st.Atimespec.Nsec = int64(attr.ATimeNanoSeconds)
	st.Atimespec.Sec = int64(attr.ATimeSeconds)
	st.Mtimespec.Nsec = int64(attr.MTimeNanoSeconds)
	st.Mtimespec.Sec = int64(attr.MTimeSeconds)
	st.Ctimespec.Nsec = int64(attr.CTimeNanoSeconds)
	st.Ctimespec.Sec = int64(attr.CTimeSeconds)
	st.Gid = uint32(attr.GID)
	st.Gen = uint32(attr.Gen)
	st.Blksize = int32(attr.BlockSize)
	st.Blocks = int64(attr.Blocks)
	st.Uid = uint32(attr.UID)
	st.Size = int64(attr.Size)
	st.Nlink = uint16(attr.NLink)
	// st.Birthtimespec.Nsec = int64(attr.BTimeNanoSeconds)
	// st.Birthtimespec.Sec = int64(attr.BTimeSeconds)
	return st
}
