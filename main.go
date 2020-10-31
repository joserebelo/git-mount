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
	"github.com/joserebelo/git-mount/git"
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
		file := &File{treeish: treeish, name: file_name, path: file_path}
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

	err = fs.Serve(c, &GitFS{root: root})
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
