package main

import (
	"syscall"

	"github.com/hugelgupf/p9/p9"
)

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
