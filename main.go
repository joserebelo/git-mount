// Hellofs implements a simple "hello world" file system.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"golang.org/x/net/context"
	"taterbase.me/git-mount/git"
)

var mountpoint string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

func cleanup() {
	err := fuse.Unmount(mountpoint)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	flag.Usage = usage
	flag.Parse()

	var err error
	mountpoint = flag.Arg(0)
	if len(mountpoint) == 0 {
		fmt.Println("No folder specified for mounting,")
		fmt.Println("creating temp directory")
		mountpoint, err = ioutil.TempDir("", "git-mount-")
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("Mounting at %s", mountpoint)
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("git-mount"),
		fuse.Subtype("git-mount"),
		fuse.LocalVolume(),
		fuse.VolumeName("katamari:3b4cfd3"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Clean up on ctrl-{c,d}
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT)
	go func() {
		<-ch
		cleanup()
		os.Exit(1)
	}()

	err = fs.Serve(c, FS{})
	if err != nil {
		log.Fatal(err)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

// FS implements the hello world file system.
type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct{}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "hello.txt" {
		return File{}, nil
	}
	return nil, fuse.ENOENT
}

var dirDirs = []fuse.Dirent{
	{Inode: 2, Name: "hello.txt", Type: fuse.DT_File},
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return dirDirs, nil
}

// File implements both Node and Handle for the hello file.
type File struct{}

const greeting = "hello, world\n"

func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(greeting))
	return nil
}

func (File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(greeting), nil
}
