package main

import (
	"syscall"

	"github.com/hugelgupf/p9/p9"
)

func sysStat(attr p9.Attr) *syscall.Stat_t {
	st := &syscall.Stat_t{}
	st.Mode = uint32(attr.Mode)
	st.Dev = uint64(attr.RDev)
	st.Atim.Nsec = int64(attr.ATimeNanoSeconds)
	st.Atim.Sec = int64(attr.ATimeSeconds)
	st.Mtim.Nsec = int64(attr.MTimeNanoSeconds)
	st.Mtim.Sec = int64(attr.MTimeSeconds)
	st.Ctim.Nsec = int64(attr.CTimeNanoSeconds)
	st.Ctim.Sec = int64(attr.CTimeSeconds)
	st.Gid = uint32(attr.GID)
	st.Blksize = int64(attr.BlockSize)
	st.Blocks = int64(attr.Blocks)
	st.Uid = uint32(attr.UID)
	st.Size = int64(attr.Size)
	st.Nlink = uint64(attr.NLink)
	return st
}
