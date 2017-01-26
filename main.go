package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cli "github.com/urfave/cli"

	core "github.com/ipfs/go-ipfs/core"
	cu "github.com/ipfs/go-ipfs/core/coreunix"
	filestore "github.com/ipfs/go-ipfs/filestore"
	balanced "github.com/ipfs/go-ipfs/importer/balanced"
	chunk "github.com/ipfs/go-ipfs/importer/chunk"
	h "github.com/ipfs/go-ipfs/importer/helpers"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"

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
	app := cli.NewApp()
	app.Commands = []cli.Command{
		makePackCommand,
		verifyPackCommand,
		repoCommand,
		servePackCommand,
	}

	app.RunAndExitOnError()
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

		go func() {
			defer close(done)
			defer manifest.Close()
			for v := range output {
				ao := v.(*cu.AddedObject)
				towrite := "." + ao.Name[len(dirname):]
				fmt.Fprintf(manifest, "%s\tFMTSTR\t%s\n", ao.Hash, towrite)
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
	Action: func(c *cli.Context) error {
		packpath := filepath.Join(cwd, PackRepo)
		if !fsrepo.IsInitialized(packpath) {
			return fmt.Errorf("TODO: autogen repo on serve")
		}

		r, err := getRepo()
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

		<-nd.Context().Done()
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

		var issue bool
		scan := bufio.NewScanner(fi)
		for scan.Scan() {
			parts := strings.SplitN(scan.Text(), "\t", 3)
			hash := parts[0]
			fmtstr := parts[1]
			path := parts[2]

			// don't use this yet
			_ = fmtstr

			fi, err := os.Open(path)
			switch {
			case os.IsNotExist(err):
				fmt.Printf("error: in manifest, missing from pack: %s\n", path)
				issue = true
				continue
			default:
				fmt.Printf("error: checking file %s: %s\n", path, err)
				issue = true
				continue
			case err == nil:
				// continue
			}

			st, err := fi.Stat()
			if err != nil {
				return err
			}
			if st.IsDir() {
				continue
			}

			spl := chunk.NewSizeSplitter(fi, chunk.DefaultBlockSize)
			params := &h.DagBuilderParams{
				Dagserv:   ds,
				NoCopy:    true,
				RawLeaves: true,
				Maxlinks:  h.DefaultLinksPerBlock,
			}
			dbh := params.New(spl)

			nd, err := balanced.BalancedLayout(dbh)
			if err != nil {
				return err
			}

			if nd.Cid().String() != hash {
				fmt.Printf("Checksum mismatch on %s. (%s)\n", path, nd.Cid().String())
				issue = true
				continue
			}
		}
		if !issue {
			fmt.Println("Pack verified successfully!")
		} else {
			fmt.Println("Pack verify found some corruption.")
		}
		return nil
	},
}
