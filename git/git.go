package git

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
)

func ListFiles(treeish, dir string) (filepaths []string, err error) {
	cmd := exec.Command(
		"git",
		"ls-tree",
		"-r",
		"--name-only",
		treeish,
		dir)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		return nil, err
	}

	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), err
}

func ShowContents(treeish, path string) (content string, err error) {
	cmd := exec.Command(
		"git",
		"show",
		treeish+":"+path)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Println(stderr.String())
		return "", err
	}

	return stdout.String(), nil
}
