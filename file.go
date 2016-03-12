package main

import (
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
	"taterbase.me/git-mount/git"
)

type File struct {
	treeish string
	name    string
	path    string
	length  uint64
}

func (f *File) IsDir() bool {
	return false
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Path() string {
	return f.path
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0444
	a.Size = f.length
	return nil
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	if !req.Flags.IsReadOnly() {
		return nil, fuse.Errno(syscall.EACCES)
	}
	resp.Flags |= fuse.OpenKeepCache
	return f, nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	content, err := git.ShowContents(f.treeish, f.Path())
	if err != nil {
		return err
	}
	fuseutil.HandleRead(req, resp, []byte(content))
	return nil
}

type GitFS struct {
	root fs.Node
}

func (gfs *GitFS) Root() (fs.Node, error) {
	return gfs.root, nil
}
