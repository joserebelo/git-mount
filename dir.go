package main

import (
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

type Dir struct {
	nodes []Node
	name  string
	path  string
}

func (d *Dir) IsDir() bool {
	return true
}

func (d *Dir) Name() string {
	return d.name
}

func (d *Dir) Path() string {
	return d.path
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | os.ModeDir | 0555
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for _, node := range d.nodes {
		if node.Name() == name {
			return node, nil
		}
	}
	return nil, fuse.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var dirents []fuse.Dirent
	for _, node := range d.nodes {
		dirent := fuse.Dirent{
			Name: node.Name(),
		}
		if node.IsDir() {
			dirent.Type = fuse.DT_Dir
		} else {
			dirent.Type = fuse.DT_File
		}
		dirents = append(dirents, dirent)
	}
	return dirents, nil
}

func (d *Dir) AddNode(node Node) {
	d.nodes = append(d.nodes, node)
}
