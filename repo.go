package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cli "github.com/urfave/cli"

	blockstore "github.com/ipfs/go-ipfs/blocks/blockstore"
	pin "github.com/ipfs/go-ipfs/pin"
	gc "github.com/ipfs/go-ipfs/pin/gc"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	cid "gx/ipfs/QmcTcsTvfaeEBRFo1TkFgT8sRmgi1n1LTZpecfVP8fzpGD/go-cid"
)

var repoCommand = cli.Command{
	Name:  "repo",
	Usage: "create (or update) a temporary ipfs object repo at '.ipfs-pack'",
	Subcommands: []cli.Command{
		repoMakeCommand,
		repoGcCommand,
		repoLsCommand,
		repoRmCommand,
	},
}
var repoMakeCommand = cli.Command{
	Name:  "make",
	Usage: "create (or update) the pack repo for this pack directory",
	Action: func(c *cli.Context) error {
		fmt.Println("not yet implemented")
		return nil
	},
}

var repoRmCommand = cli.Command{
	Name:  "rm",
	Usage: "remove this pack's ipfs repo",
	Action: func(c *cli.Context) error {
		packpath := filepath.Join(cwd, PackRepo)
		if !fsrepo.IsInitialized(packpath) {
			return fmt.Errorf("no repo found at ./.ipfs-pack")
		}

		return os.RemoveAll(packpath)
	},
}

var repoGcCommand = cli.Command{
	Name:  "gc",
	Usage: "garbage collect the pack repo",
	Action: func(c *cli.Context) error {
		packpath := filepath.Join(cwd, PackRepo)
		if !fsrepo.IsInitialized(packpath) {
			return fmt.Errorf("no repo found at ./.ipfs-pack")
		}

		fsr, err := fsrepo.Open(packpath)
		if err != nil {
			return err
		}

		bstore, ds := buildDagserv(fsr.Datastore(), fsr.FileManager())
		gcbs := blockstore.NewGCBlockstore(bstore, blockstore.NewGCLocker())
		pinner := pin.NewPinner(fsr.Datastore(), ds, ds)

		root, err := getManifestRoot()
		if err != nil {
			return err
		}

		pinner.PinWithMode(root, pin.Recursive)
		if err := pinner.Flush(); err != nil {
			return err
		}

		out, err := gc.GC(context.Background(), gcbs, ds, pinner, nil)
		if err != nil {
			return err
		}

		for k := range out {
			fmt.Printf("removed %s\n", k)
		}

		return nil
	},
}

var repoLsCommand = cli.Command{
	Name:  "ls",
	Usage: "list all cids in the pack repo",
	Action: func(c *cli.Context) error {
		packpath := filepath.Join(cwd, PackRepo)
		if !fsrepo.IsInitialized(packpath) {
			return fmt.Errorf("no repo found at ./.ipfs-pack")
		}

		fsr, err := fsrepo.Open(packpath)
		if err != nil {
			return err
		}

		bstore, _ := buildDagserv(fsr.Datastore(), fsr.FileManager())
		keys, err := bstore.AllKeysChan(context.Background())
		if err != nil {
			return err
		}

		for k := range keys {
			fmt.Println(k)
		}
		return nil
	},
}

func getManifestRoot() (*cid.Cid, error) {
	fi, err := os.Open(ManifestFilename)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, fmt.Errorf("error: no %s found", ManifestFilename)
		default:
			return nil, fmt.Errorf("error opening %s: %s", ManifestFilename, err)
		}
	}

	_, err = fi.Seek(-512, os.SEEK_END)
	if err != nil {
		return nil, err
	}

	scan := bufio.NewScanner(fi)
	var lastline string
	for scan.Scan() {
		lastline = scan.Text()
	}

	hash := strings.SplitN(lastline, "\t", 2)[0]
	return cid.Decode(hash)
}
