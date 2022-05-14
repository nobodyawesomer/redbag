package chroot

import "io/fs"

type FS interface {
	fs.ReadDirFS
	fs.ReadFileFS
}
