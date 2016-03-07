package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
	"taterbase.me/git-mount/git"
)

var (
	hash_map = make(map[string]*Dir)
	root     *Dir

	_ fs.FS                 = (*GitFS)(nil)
	_ Node                  = (*File)(nil)
	_ fs.NodeOpener         = (*File)(nil)
	_ fs.Handle             = (*File)(nil)
	_ fs.HandleReader       = (*File)(nil)
	_ Node                  = (*Dir)(nil)
	_ fs.NodeStringLookuper = (*Dir)(nil)
	_ fs.HandleReadDirAller = (*Dir)(nil)

	treeish string
)

type Node interface {
	fs.Node
	Name() string
	Path() string
	IsDir() bool
}

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

type File struct {
	name   string
	path   string
	length uint64
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
	content, err := git.ShowContents(treeish, f.Path())
	if err != nil {
		return err
	}
	fuseutil.HandleRead(req, resp, []byte(content))
	return nil
}

type GitFS struct{}

func (gfs *GitFS) Root() (fs.Node, error) {
	return root, nil
}

func main() {
	flag.Parse()
	treeish = flag.Arg(0)
	if len(treeish) == 0 {
		treeish = "HEAD"
	}

	dir := ""
	paths, err := git.ListFiles(treeish, dir)
	if err != nil {
		log.Fatalf("Unable to list files for treeish %s: %v", treeish,
			err)
	}

	root = &Dir{
		name: "/",
		path: ".",
	}

	hash_map["."] = root

	for _, file_path := range paths {
		dir, file_name := filepath.Split(file_path)
		file := &File{name: file_name, path: file_path}
		content, err := git.ShowContents(treeish, file_path)
		if err != nil {
			log.Fatal(err)
		}
		file.length = uint64(len(content))
		parent := getDirent(filepath.Dir(dir))
		parent.AddNode(file)
	}

	mountpoint, err := ioutil.TempDir("", "git-mount-")
	if err != nil {
		log.Fatal(err)
	}

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("git-mount"),
		fuse.Subtype("git-mountfs"),
		fuse.LocalVolume(),
		fuse.VolumeName(treeish),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	log.Print("mounted at: ", mountpoint)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT)
	go func() {
		<-ch
		err := fuse.Unmount(mountpoint)
		if err != nil {
			log.Print(err)
		}
		os.Exit(1)
	}()

	err = fs.Serve(c, &GitFS{})
	if err != nil {
		log.Fatal(err)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

func getDirent(path string) *Dir {
	dirent, ok := hash_map[path]
	if !ok {
		name := filepath.Base(path)
		parent_dir := filepath.Dir(path)

		parent := getDirent(parent_dir)
		dirent = &Dir{
			name: name,
			path: path,
		}
		parent.AddNode(dirent)
		hash_map[path] = dirent
	}
	return dirent
}
