package main

import (
	"encoding/binary"
	"encoding/hex"
	mh "github.com/multiformats/go-multihash"
)

const (
	DefaultChunker = iota
	DefaultRaw
	SizeChunker
	RabinChunker
)

const (
	BalancedLayout = iota
	TrickleLayout
)

const (
	DefaultImporter = iota
)

var DefaultImporterSettings = Importer{
	Version: 0,
	Which:   DefaultImporter,
	Args: ImportArgs{
		Hash:    mh.SHA2_256,
		Layout:  BalancedLayout,
		Chunker: DefaultRaw,
	},
}

type Importer struct {
	Version uint64
	Which   uint64
	Args    ImportArgs
}

type ImportArgs struct {
	Hash    uint64
	Layout  uint64
	Chunker uint64
}

func (i Importer) String() string {
	buf := make([]byte, 16)
	n := binary.PutUvarint(buf, i.Version)
	n += binary.PutUvarint(buf[n:], i.Which)
	n += binary.PutUvarint(buf[n:], i.Args.Hash)
	n += binary.PutUvarint(buf[n:], i.Args.Layout)
	n += binary.PutUvarint(buf[n:], i.Args.Chunker)

	return "f" + hex.EncodeToString(buf[:n])
}
