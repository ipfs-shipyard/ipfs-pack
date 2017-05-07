package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	blockstore "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/blocks/blockstore"
	blockservice "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/blockservice"
	files "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/commands/files"
	cu "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/core/coreunix"
	offline "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/exchange/offline"
	filestore "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/filestore"
	dag "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/merkledag"
	repo "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/repo"
	config "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/repo/config"
	fsrepo "gx/ipfs/Qmc54PqZpTxK1t5PzrZkuSzWFiw3E1RwMDuSefKwh115y1/go-ipfs/repo/fsrepo"

	ds "gx/ipfs/QmRWDav6mzWseLWeYfVd5fvUKiVe9xNH29YfMF438fG364/go-datastore"
)

func openManifestFile(workdir string) (*os.File, error) {
	fi, err := os.Open(filepath.Join(workdir, ManifestFilename))
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, fmt.Errorf("error: no %s found in %s", ManifestFilename, workdir)
		default:
			return nil, fmt.Errorf("error opening %s: %s", ManifestFilename, err)
		}
	}
	return fi, nil
}

func getRepo(workdir string) (repo.Repo, error) {
	packpath := filepath.Join(workdir, ".ipfs-pack")
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

func getFilteredDirFile(workdir string) (files.File, error) {
	contents, err := ioutil.ReadDir(workdir)
	if err != nil {
		return nil, err
	}
	dirname := filepath.Base(workdir)

	var farr []files.File
	for _, ent := range contents {
		if ent.Name() == ManifestFilename || ent.Name() == ManifestFilename+".tmp" {
			continue
		}
		if strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		f, err := files.NewSerialFile(filepath.Join(dirname, ent.Name()), filepath.Join(workdir, ent.Name()), false, ent)
		if err != nil {
			return nil, err
		}
		farr = append(farr, f)
	}

	return files.NewSliceFile(dirname, workdir, farr), nil
}
