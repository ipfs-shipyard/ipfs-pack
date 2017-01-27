package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	cli "gx/ipfs/QmVahSzvB3Upf5dAW15dpktF6PXb4z9V5LohmbcUqktyF4/cli"

	core "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/core"
	cu "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/core/coreunix"
	bitswap "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/exchange/bitswap"
	filestore "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/filestore"
	balanced "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/importer/balanced"
	chunk "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/importer/chunk"
	h "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/importer/helpers"
	dag "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/merkledag"
	fsrepo "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/repo/fsrepo"
	ft "gx/ipfs/QmQ3zzxvxdX2YGogDpx23YHKRZ4rmqGoXmnoJNdwzxtkhc/go-ipfs/unixfs"

	human "gx/ipfs/QmPSBJL4momYnE7DcUyk2DVhD6rH488ZmHBGLbxNdhU44K/go-humanize"
	ds "gx/ipfs/QmRWDav6mzWseLWeYfVd5fvUKiVe9xNH29YfMF438fG364/go-datastore"
)

var (
	cwd string
)

func init() {
	d, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cwd = d
}

const (
	ManifestFilename = "PackManifest"
	PackRepo         = ".ipfs-pack"
)

func main() {
	if err := doMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doMain() error {
	proffi := os.Getenv("IPFS_PACK_PROFILE")
	if proffi != "" {
		fi, err := os.Create(proffi)
		if err != nil {
			return err
		}

		defer fi.Close()
		err = pprof.StartCPUProfile(fi)
		if err != nil {
			return err
		}

		defer pprof.StopCPUProfile()
	}

	app := cli.NewApp()
	app.Commands = []cli.Command{
		makePackCommand,
		verifyPackCommand,
		repoCommand,
		servePackCommand,
	}

	return app.Run(os.Args)
}

var makePackCommand = cli.Command{
	Name:  "make",
	Usage: "makes the package, overwriting the PackManifest file.",
	Action: func(c *cli.Context) error {
		repo, err := getRepo()
		if err != nil {
			return err
		}

		adder, err := getAdder(repo.Datastore(), repo.FileManager())
		if err != nil {
			return err
		}
		dirname := filepath.Base(cwd)

		output := make(chan interface{})
		adder.Out = output

		done := make(chan struct{})
		manifest, err := os.Create(ManifestFilename)
		if err != nil {
			return err
		}

		imp := DefaultImporterSettings.String()

		go func() {
			defer close(done)
			defer manifest.Close()
			for v := range output {
				ao := v.(*cu.AddedObject)
				towrite := "." + ao.Name[len(dirname):]
				fmt.Fprintf(manifest, "%s\t%s\t%s\n", ao.Hash, imp, towrite)
			}
		}()

		sf, err := getFilteredDirFile()
		if err != nil {
			return err
		}

		err = adder.AddFile(sf)
		if err != nil {
			return err
		}

		_, err = adder.Finalize()
		if err != nil {
			return err
		}

		close(output)
		<-done
		fmt.Println("wrote PackManifest")

		return nil
	},
}

var servePackCommand = cli.Command{
	Name:  "serve",
	Usage: "start an ipfs node to serve this pack's contents",
	Flags: []cli.Flag{
		cli.BoolTFlag{
			Name:  "verify",
			Usage: "verify integrity of pack before serving",
		},
	},
	Action: func(c *cli.Context) error {
		packpath := filepath.Join(cwd, PackRepo)
		if !fsrepo.IsInitialized(packpath) {
			return fmt.Errorf("TODO: autogen repo on serve")
		}

		r, err := getRepo()
		if err != nil {
			return err
		}

		verify := c.BoolT("verify")
		if verify {
			_, ds := buildDagserv(r.Datastore(), r.FileManager())
			fi, err := os.Open(ManifestFilename)
			if err != nil {
				switch {
				case os.IsNotExist(err):
					return fmt.Errorf("error: no %s found", ManifestFilename)
				default:
					return fmt.Errorf("error opening %s: %s", ManifestFilename, err)
				}
			}
			defer fi.Close()

			problem, err := verifyPack(ds, fi)
			if err != nil {
				return err
			}

			if problem {
				return fmt.Errorf("pack verify failed, refusing to serve")
			} else {
				fmt.Println("verified pack, starting server...")
			}
		}

		root, err := getManifestRoot()
		if err != nil {
			return err
		}

		cfg := &core.BuildCfg{
			Online: true,
			Repo:   r,
		}

		nd, err := core.NewNode(context.Background(), cfg)
		if err != nil {
			return err
		}
		fmt.Println("Serving data in this pack...")
		fmt.Printf("Peer ID: %s\n", nd.Identity.Pretty())
		for _, a := range nd.PeerHost.Addrs() {
			fmt.Printf("    %s\n", a)
		}

		fmt.Printf("Pack root is %s\n", root)
		tick := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-nd.Context().Done():
				return nil
			case <-tick.C:
				st, err := nd.Exchange.(*bitswap.Bitswap).Stat()
				if err != nil {
					fmt.Println("error getting block stat: ", err)
					continue
				}
				fmt.Printf(strings.Repeat("    ", 12) + "\r")
				fmt.Printf("Shared: %6d blocks, %s total data uploaded\r", st.BlocksSent, human.Bytes(st.DataSent))
			}
		}

		return nil
	},
}

var verifyPackCommand = cli.Command{
	Name:  "verify",
	Usage: "verifies the ipfs-pack manifest file is correct",
	Action: func(c *cli.Context) error {
		// TODO: check for files in pack that arent in manifest
		fi, err := os.Open(ManifestFilename)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				return fmt.Errorf("error: no %s found", ManifestFilename)
			default:
				return fmt.Errorf("error opening %s: %s", ManifestFilename, err)
			}
		}

		var dstore ds.Batching
		var fm *filestore.FileManager
		packpath := filepath.Join(cwd, ".ipfs-pack")
		if fsrepo.IsInitialized(packpath) {
			r, err := getRepo()
			if err != nil {
				return err
			}
			dstore = r.Datastore()
			fm = r.FileManager()
		} else {
			dstore = ds.NewNullDatastore()
		}

		_, ds := buildDagserv(dstore, fm)

		issue, err := verifyPack(ds, fi)
		if err != nil {
			return err
		}

		if !issue {
			fmt.Println("Pack verified successfully!")
		} else {
			fmt.Println("Pack verify found some corruption.")
		}
		return nil
	},
}

func verifyPack(ds dag.DAGService, manif io.Reader) (bool, error) {
	var issue bool
	scan := bufio.NewScanner(manif)

	imp := DefaultImporterSettings.String()
	for scan.Scan() {
		parts := strings.SplitN(scan.Text(), "\t", 3)
		hash := parts[0]
		fmtstr := parts[1]
		path := parts[2]

		if fmtstr != imp {
			fmt.Printf("error: unsupported importer settings in manifest file: %s\n", fmtstr)
			issue = true
			continue
		}

		params := &h.DagBuilderParams{
			Dagserv:   ds,
			NoCopy:    true,
			RawLeaves: true,
			Maxlinks:  h.DefaultLinksPerBlock,
		}

		ok, err := verifyItem(path, hash, params)
		if err != nil {
			return false, err
		}
		if !ok {
			issue = true
		}
	}
	return issue, nil
}

func verifyItem(path, hash string, params *h.DagBuilderParams) (bool, error) {
	st, err := os.Lstat(path)
	switch {
	case os.IsNotExist(err):
		fmt.Printf("error: in manifest, missing from pack: %s\n", path)
		return false, nil
	default:
		fmt.Printf("error: checking file %s: %s\n", path, err)
		return false, nil
	case err == nil:
		// continue
	}

	if st.IsDir() {
		return true, nil
	}
	if st.Mode()&os.ModeSymlink != 0 {
		trgt, err := os.Readlink(path)
		if err != nil {
			return false, err
		}

		data, err := ft.SymlinkData(trgt)
		if err != nil {
			return false, err
		}

		nd := new(dag.ProtoNode)
		nd.SetData(data)
		if nd.Cid().String() != hash {
			fmt.Printf("Checksum mismatch on symlink: %s", path)
			return false, nil
		}
		return true, nil
	}

	fi, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer fi.Close()

	spl := chunk.NewSizeSplitter(fi, chunk.DefaultBlockSize)
	dbh := params.New(spl)

	nd, err := balanced.BalancedLayout(dbh)
	if err != nil {
		return false, err
	}

	if nd.Cid().String() != hash {
		fmt.Printf("Checksum mismatch on %s. (%s)\n", path, nd.Cid().String())
		return false, nil
	}
	return true, nil
}
