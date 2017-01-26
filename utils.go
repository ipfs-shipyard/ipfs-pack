package main

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"

	blockstore "github.com/ipfs/go-ipfs/blocks/blockstore"
	blockservice "github.com/ipfs/go-ipfs/blockservice"
	files "github.com/ipfs/go-ipfs/commands/files"
	cu "github.com/ipfs/go-ipfs/core/coreunix"
	offline "github.com/ipfs/go-ipfs/exchange/offline"
	filestore "github.com/ipfs/go-ipfs/filestore"
	dag "github.com/ipfs/go-ipfs/merkledag"
	repo "github.com/ipfs/go-ipfs/repo"
	config "github.com/ipfs/go-ipfs/repo/config"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

	ds "gx/ipfs/QmRWDav6mzWseLWeYfVd5fvUKiVe9xNH29YfMF438fG364/go-datastore"
)

func getRepo() (repo.Repo, error) {
	packpath := filepath.Join(cwd, ".ipfs-pack")
	if !fsrepo.IsInitialized(packpath) {
		cfg, err := config.Init(ioutil.Discard, 1024)
		if err != nil {
			return nil, err
		}

		cfg.Addresses.API = ""
		cfg.Addresses.Gateway = "/ip4/127.0.0.1/tcp/0"
		cfg.Addresses.Swarm = []string{"/ip4/0.0.0.0/tcp/0"}
		cfg.Datastore.NoSync = true
		cfg.Experimental.FilestoreEnabled = true
		cfg.Reprovider.Interval = "0"

		err = fsrepo.Init(packpath, cfg)
		if err != nil {
			return nil, err
		}
	}
	return fsrepo.Open(packpath)
}

func buildDagserv(dstore ds.Batching, fm *filestore.FileManager) (blockstore.Blockstore, dag.DAGService) {
	var bs blockstore.Blockstore = blockstore.NewBlockstore(dstore)
	if fm != nil {
		bs = filestore.NewFilestore(bs, fm)
	}
	bserv := blockservice.New(bs, offline.Exchange(bs))
	return bs, dag.NewDAGService(bserv)
}

func getAdder(dstore ds.Batching, fm *filestore.FileManager) (*cu.Adder, error) {
	bstore, dserv := buildDagserv(dstore, fm)

	gcbs := blockstore.NewGCBlockstore(bstore, blockstore.NewGCLocker())
	adder, err := cu.NewAdder(context.Background(), nil, gcbs, dserv)
	if err != nil {
		return nil, err
	}
	adder.NoCopy = true
	adder.RawLeaves = true
	return adder, nil
}

func getFilteredDirFile() (files.File, error) {
	contents, err := ioutil.ReadDir(cwd)
	if err != nil {
		return nil, err
	}
	dirname := filepath.Base(cwd)

	var farr []files.File
	for _, ent := range contents {
		if ent.Name() == ManifestFilename {
			continue
		}
		if strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		f, err := files.NewSerialFile(filepath.Join(dirname, ent.Name()), filepath.Join(cwd, ent.Name()), false, ent)
		if err != nil {
			return nil, err
		}
		farr = append(farr, f)
	}

	return files.NewSliceFile(dirname, cwd, farr), nil
}
