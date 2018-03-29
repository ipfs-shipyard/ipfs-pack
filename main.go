package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"time"

	cli "gx/ipfs/QmVahSzvB3Upf5dAW15dpktF6PXb4z9V5LohmbcUqktyF4/cli"

	cid "gx/ipfs/QmNp85zy9RLrQ5oQD4hPyS39ezrrXpcaa7R4Y9kxdWQLLQ/go-cid"
	files "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/commands/files"
	core "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/core"
	cu "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/core/coreunix"
	bitswap "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/exchange/bitswap"
	filestore "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/filestore"
	balanced "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/importer/balanced"
	chunk "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/importer/chunk"
	h "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/importer/helpers"
	dag "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/merkledag"
	fsrepo "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/repo/fsrepo"
	ft "gx/ipfs/QmdKL1GVaUaDVt3JUWiYQSLYRsJMym2KRWxsiXAeEU6pzX/go-ipfs/unixfs"

	node "gx/ipfs/QmPN7cwmpcc4DWXb4KTB9dNAJgjuPY69h3npsMfhRrQL9c/go-ipld-format"
	human "gx/ipfs/QmPSBJL4momYnE7DcUyk2DVhD6rH488ZmHBGLbxNdhU44K/go-humanize"
	ds "gx/ipfs/QmVSase1JP7cq9QkPT46oNwdp9pT6kBkG3oqS14y3QcZjG/go-datastore"

	pb "gx/ipfs/QmeWjRodbcZFKe5tMN7poEx3izym6osrLSnTLf9UjJZBbs/pb"
)

const PackVersion = "v0.6.0"

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

func setupProfiling() (func(), error) {
	halt := func() {}

	proffi := os.Getenv("IPFS_PACK_CPU_PROFILE")
	if proffi != "" {
		fi, err := os.Create(proffi)
		if err != nil {
			return nil, err
		}

		err = pprof.StartCPUProfile(fi)
		if err != nil {
			return nil, err
		}

		halt = func() {
			pprof.StopCPUProfile()
			fi.Close()
		}
	}

	memprofi := os.Getenv("IPFS_PACK_MEM_PROFILE")
	if memprofi != "" {
		go func() {
			for range time.NewTicker(time.Second * 5).C {
				fi, err := os.Create(memprofi)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error writing heap profile: %s", err)
					return
				}

				if err := pprof.WriteHeapProfile(fi); err != nil {
					fmt.Fprintf(os.Stderr, "error writing heap profile: %s", err)
					return
				}

				fi.Close()
			}
		}()
	}

	return halt, nil
}

func doMain() error {
	if haltprof, err := setupProfiling(); err != nil {
		return err
	} else {
		defer haltprof()
	}

	app := cli.NewApp()
	app.Usage = "A filesystem packing tool"
	app.Version = PackVersion
	app.Commands = []cli.Command{
		makePackCommand,
		verifyPackCommand,
		repoCommand,
		servePackCommand,
	}

	return app.Run(os.Args)
}

var makePackCommand = cli.Command{
	Name:      "make",
	Usage:     "makes the package, overwriting the PackManifest file",
	ArgsUsage: "<dir>",
	Action: func(c *cli.Context) error {
		workdir := cwd
		if c.Args().Present() {
			argpath, err := filepath.Abs(c.Args().First())
			if err != nil {
				return err
			}
			workdir = argpath
		}

		repo, err := getRepo(workdir)
		if err != nil {
			return err
		}

		adder, err := getAdder(repo.Datastore(), repo.FileManager())
		if err != nil {
			return err
		}
		dirname := filepath.Base(workdir)

		output := make(chan interface{})
		adder.Out = output
		adder.Progress = true

		done := make(chan struct{})
		manifestName := filepath.Join(workdir, ManifestFilename)
		manifest, err := os.Create(manifestName + ".tmp")
		if err != nil {
			return err
		}

		imp := DefaultImporterSettings.String()

		fmt.Println("Building IPFS Pack")

		bar := pb.New64(-1)
		bar.Units = pb.U_BYTES
		bar.ShowSpeed = true
		bar.ShowPercent = false
		bar.Start()

		go func() {
			defer close(done)
			defer manifest.Close()
			var sizetotal int64
			var sizethis int64
			for v := range output {
				ao := v.(*cu.AddedObject)
				if ao.Bytes == 0 {
					sizetotal += sizethis
					sizethis = 0
				} else {
					sizethis = ao.Bytes
					bar.Set64(sizetotal + sizethis)
				}
				if ao.Hash == "" {
					continue
				}
				towrite := ao.Name[len(dirname):]
				if len(towrite) > 0 {
					towrite = towrite[1:]
				} else {
					towrite = "."
				}
				fmt.Fprintf(manifest, "%s\t%s\t%s\n", ao.Hash, imp, escape(towrite))
			}
		}()

		sf, err := getFilteredDirFile(workdir)
		if err != nil {
			return err
		}

		go func() {
			sizer := sf.(files.SizeFile)
			size, err := sizer.Size()
			if err != nil {
				fmt.Println("warning: could not compute size:", err)
				return
			}
			bar.Total = size
			bar.ShowPercent = true
		}()

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
		err = os.Rename(manifestName+".tmp", manifestName)
		if err != nil {
			fmt.Printf("Pack creation completed sucessfully, but we failed to rename '%s' to '%s' due to the following error:\n", manifestName+".tmp", manifestName)
			fmt.Println(err)
			fmt.Println("To resolve the issue, manually rename the mentioned file.")
			os.Exit(1)
		}

		mes := "wrote PackManifest"
		clearBar(bar, mes)
		return nil
	},
}

func clearBar(bar *pb.ProgressBar, mes string) {
	fmt.Printf("\r%s%s\n", mes, strings.Repeat(" ", bar.GetWidth()-len(mes)))
}
func getPackRoot(nd *core.IpfsNode, workdir string) (node.Node, error) {
	root, err := getManifestRoot(workdir)
	if err != nil {
		return nil, err
	}

	proot, err := nd.DAG.Get(context.Background(), root)
	if err != nil {
		return nil, err
	}

	pfi, err := os.Open(filepath.Join(workdir, ManifestFilename))
	if err != nil {
		return nil, err
	}

	manifhash, err := cu.Add(nd, pfi)
	if err != nil {
		return nil, err
	}

	manifcid, err := cid.Decode(manifhash)
	if err != nil {
		return nil, err
	}

	manifnode, err := nd.DAG.Get(context.Background(), manifcid)
	if err != nil {
		return nil, err
	}

	prootpb := proot.(*dag.RawNode)
	prootpb.AddNodeLinkClean(ManifestFilename, manifnode.(*dag.RawNode))
	_, err = nd.DAG.Add(prootpb)
	if err != nil {
		return nil, err
	}
	return prootpb, nil
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
		workdir := cwd
		if c.Args().Present() {
			argpath, err := filepath.Abs(c.Args().First())
			if err != nil {
				return err
			}
			workdir = argpath
		}
		if _, err := os.Stat(workdir); os.IsNotExist(err) {
			fmt.Printf("No such directory: '%s'\n\nCOMMAND HELP:\n", workdir)
			return cli.ShowCommandHelp(c, "serve")
		}

		packpath := filepath.Join(workdir, PackRepo)
		if !fsrepo.IsInitialized(packpath) {
			return fmt.Errorf("No ipfs-pack found in '%s'\nplease run 'ipfs-pack make' before 'ipfs-pack serve'", cwd)
		}

		r, err := getRepo(workdir)
		if err != nil {
			return fmt.Errorf("error opening repo: %s", err)
		}

		verify := c.BoolT("verify")
		if verify {
			_, ds := buildDagserv(r.Datastore(), r.FileManager())
			fi, err := os.Open(filepath.Join(workdir, ManifestFilename))
			if err != nil {
				switch {
				case os.IsNotExist(err):
					return fmt.Errorf("error: no %s found in %s", ManifestFilename, workdir)
				default:
					return fmt.Errorf("error opening %s: %s", ManifestFilename, err)
				}
			}
			defer fi.Close()

			fmt.Print("Verifying pack contents before serving...")

			problem, err := verifyPack(ds, workdir, fi)
			if err != nil {
				return fmt.Errorf("error verifying pack: %s", err)
			}

			if problem {
				fmt.Println()
				return fmt.Errorf(`Pack verify failed, refusing to serve.
  If you meant to change the files, re-run 'ipfs-pack make' to rebuild the manifest
  Otherwise, replace the bad files with the originals and run 'ipfs-pack serve' again`)
			} else {
				fmt.Printf("\r" + QClrLine + "Verified pack, starting server...")
			}
		}

		cfg := &core.BuildCfg{
			Online:  true,
			Repo:    r,
			Routing: core.DHTClientOption,
		}

		nd, err := core.NewNode(context.Background(), cfg)
		if err != nil {
			return err
		}

		proot, err := getPackRoot(nd, workdir)
		if err != nil {
			return err
		}

		totsize, err := proot.(*dag.RawNode).Size()
		if err != nil {
			return err
		}

		fmt.Print(QReset)
		putMessage(1, color(Cyan, "ipfs-pack"))
		padPrint(3, "Pack Status", color(Green, "serving globally"))
		padPrint(4, "Uptime", "0m")
		padPrint(5, "Version", PackVersion)
		padPrint(6, "Pack Size", human.Bytes(totsize))
		padPrint(7, "Connections", "0 peers")
		padPrint(8, "PeerID", nd.Identity.Pretty())
		padPrint(10, "Shared", "blocks      total up    rate up")
		padPrint(11, "", "0            0           0")

		padPrint(13, "Pack Root Hash", fmt.Sprintf("dweb:%s", color(Blue, "/ipfs/"+proot.Cid().String())))

		addrs := nd.PeerHost.Addrs()
		putMessage(15, "Addresses")
		for i, a := range addrs {
			putMessage(16+i, a.String()+"/ipfs/"+nd.Identity.Pretty())
		}

		bottom := 16 + len(addrs)
		lg := NewLog(bottom+3, 10)
		putMessage(bottom+1, "Activity Log")
		putMessage(bottom+2, "------------")
		putMessage(bottom+15, "[Press Ctrl+c to shutdown]")

		tick := time.NewTicker(time.Second)
		provdelay := time.After(time.Second * 5)
		start := time.Now()
		addlog := make(chan string, 16)
		nd.PeerHost.Network().Notify(&LogNotifee{addlog})
		killed := make(chan os.Signal)
		signal.Notify(killed, os.Interrupt)

		var provinprogress bool
		for {
			putMessage(bottom+13, "")
			select {
			case <-nd.Context().Done():
				return nil
			case <-tick.C:
				npeers := len(nd.PeerHost.Network().Peers())
				padPrint(7, "Connections", fmt.Sprint(npeers)+" peers")

				st, err := nd.Exchange.(*bitswap.Bitswap).Stat()
				if err != nil {
					fmt.Println("error getting block stat: ", err)
					continue
				}

				bw := nd.Reporter.GetBandwidthTotals()
				printDataSharedLine(11, st.BlocksSent, bw.TotalOut, bw.RateOut)
				printTime(4, start)
			case <-provdelay:
				if !provinprogress {
					lg.Add("announcing pack content to the network")
					done := make(chan time.Time)
					provdelay = done
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
						err := nd.Routing.Provide(ctx, proot.Cid(), true)
						if err != nil {
							lg.Add(fmt.Sprintf("error notifying network about our pack: %s", err))
						}
						close(done)
						cancel()
					}()
					provinprogress = true
				} else {
					lg.Add("completed network announcement!")
					provdelay = time.After(time.Hour * 12)
					provinprogress = false
				}
			case mes := <-addlog:
				lg.Add(mes)
				lg.Print()
			case <-killed:
				fmt.Println(QReset)
				fmt.Println("Shutting down ipfs-pack node...")
				nd.Close()
				return nil
			}
		}

		return nil
	},
}

func printTime(line int, start time.Time) {
	t := time.Since(start)
	h := int(t.Hours())
	m := int(t.Minutes()) % 60
	s := int(t.Seconds()) % 60
	padPrint(line, "Uptime", fmt.Sprintf("%dh %dm %ds", h, m, s))
}

var verifyPackCommand = cli.Command{
	Name:  "verify",
	Usage: "verifies the ipfs-pack manifest file is correct.",
	Action: func(c *cli.Context) error {
		workdir := cwd
		if c.Args().Present() {
			argpath, err := filepath.Abs(c.Args().First())
			if err != nil {
				return err
			}
			workdir = argpath
		}

		// TODO: check for files in pack that arent in manifest
		fi, err := os.Open(filepath.Join(workdir, ManifestFilename))
		if err != nil {
			switch {
			case os.IsNotExist(err):
				return fmt.Errorf("error: no %s found in %s", ManifestFilename, workdir)
			default:
				return fmt.Errorf("error opening %s: %s", ManifestFilename, err)
			}
		}

		var dstore ds.Batching
		var fm *filestore.FileManager
		packpath := filepath.Join(workdir, ".ipfs-pack")
		if fsrepo.IsInitialized(packpath) {
			r, err := getRepo(workdir)
			if err != nil {
				return err
			}
			dstore = r.Datastore()
			fm = r.FileManager()
		} else {
			dstore = ds.NewNullDatastore()
		}

		_, ds := buildDagserv(dstore, fm)

		issue, err := verifyPack(ds, workdir, fi)
		if err != nil {
			return err
		}

		if !issue {
			fmt.Println("pack verification succeeded") // nicola is an easter egg
		} else {
			return fmt.Errorf("error: pack verification failed")
		}
		return nil
	},
}

func verifyPack(ds dag.DAGService, workdir string, manif io.Reader) (bool, error) {
	imp := DefaultImporterSettings.String()

	var issue bool
	scan := bufio.NewScanner(manif)
	for scan.Scan() {
		parts := strings.SplitN(scan.Text(), "\t", 3)
		hash := parts[0]
		fmtstr := parts[1]
		path, err := unescape(parts[2])
		if err != nil {
			fmt.Printf("%v\n", err)
			issue = true
			continue
		}

		if fmtstr != imp {
			if !issue {
				fmt.Println()
			}
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

		ok, mes, err := verifyItem(path, hash, workdir, params)
		if err != nil {
			return false, err
		}
		if !ok {
			if !issue {
				fmt.Println()
			}
			fmt.Println(mes)
			issue = true
			continue
		}
	}
	return issue, nil
}

func verifyItem(path, hash, workdir string, params *h.DagBuilderParams) (bool, string, error) {
	st, err := os.Lstat(filepath.Join(workdir, path))
	switch {
	case os.IsNotExist(err):
		return false, fmt.Sprintf("error: item in manifest, missing from pack: %s", path), nil
	default:
		return false, fmt.Sprintf("error: checking file %s: %s", path, err), nil
	case err == nil:
		// continue
	}

	if st.IsDir() {
		return true, "", nil
	}

	nd, err := addItem(filepath.Join(workdir, path), st, params)
	if err != nil {
		return false, "", err
	}

	if nd.Cid().String() != hash {
		s := fmt.Sprintf("error: checksum mismatch on %s. (%s)", path, nd.Cid().String())
		return false, s, nil
	}
	return true, "", nil
}

func addItem(path string, st os.FileInfo, params *h.DagBuilderParams) (node.Node, error) {
	if st.Mode()&os.ModeSymlink != 0 {
		trgt, err := os.Readlink(path)
		if err != nil {
			return nil, err
		}

		data, err := ft.SymlinkData(trgt)
		if err != nil {
			return nil, err
		}

		nd := new(dag.RawNode)
		nd.SetData(data)
		return nd, nil
	}

	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	rf, err := files.NewReaderPathFile(filepath.Base(path), path, fi, st)
	if err != nil {
		return nil, err
	}

	spl := chunk.NewSizeSplitter(rf, chunk.DefaultBlockSize)
	dbh := params.New(spl)

	nd, err := balanced.BalancedLayout(dbh)
	if err != nil {
		return nil, err
	}

	return nd, nil
}
