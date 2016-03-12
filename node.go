package main

import (
	"bazil.org/fuse/fs"
)

type Node interface {
	fs.Node
	Name() string
	Path() string
	IsDir() bool
}
