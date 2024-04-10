package main

import (
	"log"
	"os"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hugelgupf/p9/linux"
	"github.com/hugelgupf/p9/p9"
)

func fuseMode(t p9.QIDType) uint32 {
	switch t {
	case p9.TypeDir:
		return fuse.S_IFDIR
	case p9.TypeRegular:
		return fuse.S_IFREG
	case p9.TypeLink, p9.TypeSymlink:
		return fuse.S_IFLNK
	default:
		log.Panicf("unsupported QID<->FUSE type: %v", t)
		return 0
	}
}

func openFlags(flags uint32) p9.OpenFlags {
	accessMode := flags & syscall.O_ACCMODE
	switch accessMode {
	case syscall.O_RDONLY:
		return p9.ReadOnly
	case syscall.O_WRONLY:
		return p9.WriteOnly
	case syscall.O_RDWR:
		return p9.ReadWrite
	default:
		log.Panicf("unsupported access mode: %v", accessMode)
		return 0
	}
}

func sysErrno(err error) syscall.Errno {
	log.Println("ERR:", err)
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
		return sysErrno(t.Err)
	case *os.LinkError:
		return sysErrno(t.Err)
	}
	log.Println("!! unsupported error type:", err)
	return syscall.EINVAL
}

func sysStat(attr p9.Attr) *syscall.Stat_t {
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
	st.Birthtimespec.Nsec = int64(attr.BTimeNanoSeconds)
	st.Birthtimespec.Sec = int64(attr.BTimeSeconds)
	return st
}
