package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cli "gx/ipfs/QmVahSzvB3Upf5dAW15dpktF6PXb4z9V5LohmbcUqktyF4/cli"

	blockstore "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/blocks/blockstore"
	h "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/importer/helpers"
	mfs "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/mfs"
	pin "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/pin"
	gc "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/pin/gc"
	fsrepo "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/repo/fsrepo"
	ft "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/unixfs"

	cid "gx/ipfs/QmNp85zy9RLrQ5oQD4hPyS39ezrrXpcaa7R4Y9kxdWQLLQ/go-cid"
)

var repoCommand = cli.Command{
	Name:  "repo",
	Usage: "manipulate the ipfs repo associated with this pack",
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
		workdir := cwd
		if c.Args().Present() {
			argpath, err := filepath.Abs(c.Args().First())
			if err != nil {
				return err
			}
			workdir = argpath
		}

		fi, err := openManifestFile(workdir)
		if err != nil {
			return err
		}

		r, err := getRepo(workdir)
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
			path, err := unescape(parts[2])
			if err != nil {
				return err
			}

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

			nd, err := addItem(filepath.Join(workdir, path), st, params)
			if err != nil {
				return err
			}

			if hash != nd.Cid().String() {
				return fmt.Errorf("error: checksum fail on %s (exp %s, got %s)", path, hash, nd.Cid())
			}

			err = mfs.Mkdir(root, filepath.Dir(path), mfs.MkdirOpts{Mkparents: true})
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
	Usage: "garbage collect the pack's ipfs repo",
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

		root, err := getManifestRoot(cwd)
		if err != nil {
			return err
		}

		pinner.PinWithMode(root, pin.Recursive)
		if err := pinner.Flush(); err != nil {
			return err
		}

		out := gc.GC(context.Background(), gcbs, ds, pinner, nil)
		if err != nil {
			return err
		}

		for k := range out {
			if k.Error != nil {
				fmt.Printf("GC Error: %s\n", k.Error)
			} else {
				fmt.Printf("removed %s\n", k.KeyRemoved)
			}
		}

		return nil
	},
}

var repoLsCommand = cli.Command{
	Name:  "ls",
	Usage: "list all cids in the pack's ipfs repo",
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

func getManifestRoot(workdir string) (*cid.Cid, error) {
	fi, err := os.Open(filepath.Join(workdir, ManifestFilename))
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, fmt.Errorf("error: no %s found in %s", ManifestFilename, workdir)
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
