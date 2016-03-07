package main

import (
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
	"taterbase.me/git-mount/git"
)

var (
	hash_map = make(map[string]*DirEnt)
	root     *DirEnt
)

type DirEnt struct {
	INodes []*DirEnt
	Name   string
	Path   string
}

var _ fs.Node = (*DirEnt)(nil)

func (d *DirEnt) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | os.ModeDir | 0555
	return nil
}

type File struct {
	length uint64
}

var _ fs.Node = (*File)(nil)

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0444
	a.Size = f.length
	return nil
}

type GitFS struct{}

var _ fs.FS = (*GitFS)(nil)

func (gfs *GitFS) Root() (fs.Node, error) {
	return root, nil
}

func main() {
	treeish := "HEAD"
	dir := ""
	paths, err := git.ListFiles(treeish, dir)
	if err != nil {
		log.Fatalf("Unable to list files for treeish %s: %v", treeish,
			err)
	}

	root = &DirEnt{
		Name: "/",
		Path: ".",
	}

	hash_map["."] = root

	for _, file_path := range paths {
		dir, file_name := filepath.Split(file_path)
		file := &DirEnt{Name: file_name, Path: file_path}
		parent := getDirent(dir)
		parent.INodes = append(parent.INodes, file)
	}
	log.Printf("%+v", hash_map)
}

func getDirent(path string) *DirEnt {
	dirent, ok := hash_map[path]
	if !ok {
		name := filepath.Base(path)
		parent_dir := filepath.Dir(path)

		parent := getDirent(parent_dir)
		dirent = &DirEnt{
			Name:   name,
			Path:   path,
			INodes: []*DirEnt{},
		}
		parent.INodes = append(parent.INodes, dirent)
		hash_map[path] = dirent
	}
	return dirent
}
