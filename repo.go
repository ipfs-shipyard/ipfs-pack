package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cli "gx/ipfs/QmVahSzvB3Upf5dAW15dpktF6PXb4z9V5LohmbcUqktyF4/cli"

	blockstore "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/blocks/blockstore"
	h "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/importer/helpers"
	mfs "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/mfs"
	pin "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/pin"
	gc "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/pin/gc"
	fsrepo "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/repo/fsrepo"
	ft "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/unixfs"

	cid "gx/ipfs/QmcTcsTvfaeEBRFo1TkFgT8sRmgi1n1LTZpecfVP8fzpGD/go-cid"
)

var repoCommand = cli.Command{
	Name:  "repo",
	Usage: "manipulate the ipfs repo cache associated with this pack",
	Subcommands: []cli.Command{
		repoRegenCommand,
		repoGcCommand,
		repoLsCommand,
		repoRmCommand,
	},
}

var repoRegenCommand = cli.Command{
	Name:  "regen",
	Usage: "regenerate ipfs-pack repo for this pack",
	Action: func(c *cli.Context) error {
		fi, err := os.Open(ManifestFilename)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				return fmt.Errorf("error: no %s found", ManifestFilename)
			default:
				return fmt.Errorf("error opening %s: %s", ManifestFilename, err)
			}
		}

		r, err := getRepo()
		if err != nil {
			return err
		}

		_, dserv := buildDagserv(r.Datastore(), r.FileManager())
		root, err := mfs.NewRoot(context.Background(), dserv, ft.EmptyDirNode(), nil)
		if err != nil {
			return err
		}

		defaultFmts := DefaultImporterSettings.String()
		scan := bufio.NewScanner(fi)
		for scan.Scan() {
			parts := strings.SplitN(scan.Text(), "\t", 3)
			hash := parts[0]
			fmts := parts[1]
			path := parts[2]

			if fmts != defaultFmts {
				return fmt.Errorf("error: unsupported import settings string: %s != %s", fmts, defaultFmts)
			}

			params := &h.DagBuilderParams{
				Dagserv:   dserv,
				NoCopy:    true,
				RawLeaves: true,
				Maxlinks:  h.DefaultLinksPerBlock,
			}

			st, err := os.Lstat(path)
			switch {
			case os.IsNotExist(err):
				return fmt.Errorf("error: in manifest, missing from pack: %s\n", path)
			default:
				return fmt.Errorf("error: reading file %s: %s\n", path, err)
			case err == nil:
				// continue
			}

			if st.IsDir() {
				// TODO: maybe check that the mfs root records this as being correct?
				continue
			}

			nd, err := addItem(path, st, params)
			if err != nil {
				return err
			}

			if hash != nd.Cid().String() {
				return fmt.Errorf("error: checksum fail on %s (exp %s, got %s)", path, hash, nd.Cid())
			}

			err = mfs.Mkdir(root, filepath.Dir(path), true, false)
			if err != nil {
				return fmt.Errorf("error reconstructing tree: %s", err)
			}

			err = mfs.PutNode(root, filepath.Clean(path), nd)
			if err != nil {
				return fmt.Errorf("error adding tree node: %s", err)
			}
		}

		nd, err := root.GetValue().GetNode()
		if err != nil {
			return err
		}
		_ = nd
		fmt.Println("ipfs pack repo successfully regenerated.")

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

	st, err := fi.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() > 1024 {
		_, err = fi.Seek(-512, os.SEEK_END)
		if err != nil {
			return nil, err
		}
	}

	scan := bufio.NewScanner(fi)
	var lastline string
	for scan.Scan() {
		lastline = scan.Text()
	}

	hash := strings.SplitN(lastline, "\t", 2)[0]
	return cid.Decode(hash)
}
