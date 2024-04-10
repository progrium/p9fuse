# p9fuse

A FUSE filesystem for accessing 9P filesystems written in Go.

## Install

Make sure FUSE (or MacFUSE) is installed, as well as Go for now:

```
go install github.com/progrium/p9fuse
```

## Usage

```
p9fuse [-debug] <9PADDR> <MOUNTPOINT>
```

Use `umount` on the mountpoint to unmount and `p9fuse` will exit.

## Notes

* Currently provides a read-only filesystem from a 9P endpoint
* The 9P server will need to speak 9P2000.L as supported by [github.com/hugelgupf/p9](https://github.com/hugelgupf/p9)

## License

MIT