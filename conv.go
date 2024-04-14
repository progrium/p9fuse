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
	if (t & p9.TypeDir) == p9.TypeDir {
		return fuse.S_IFDIR
	}
	if (t & p9.TypeRegular) == p9.TypeRegular {
		return fuse.S_IFREG
	}
	if (t & p9.TypeLink) == p9.TypeLink {
		return fuse.S_IFLNK
	}
	if (t & p9.TypeSymlink) == p9.TypeSymlink {
		return fuse.S_IFLNK
	}
	log.Panicf("unsupported QID<->FUSE type: %v", t)
	return 0
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
