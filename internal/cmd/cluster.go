package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"

	dg "github.com/Donders-Institute/hpc-cluster-tools/internal/datagetter"
	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var xml bool

const (
	gib float64 = 1024 * 1024 * 1024
)

// variable may be set at the build time to fix the default location for the TorqueHelper server certificate.
var defTorqueHelperCert string

func init() {

	qstatCmd.Flags().BoolVarP(&xml, "xml", "x", false, "XML output")

	clusterCmd.PersistentFlags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	clusterCmd.PersistentFlags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")
	clusterCmd.PersistentFlags().StringVarP(&TorqueHelperCert, "cert", "c", defTorqueHelperCert, "Torque helper service certificate")

	nodeCmd.AddCommand(nodeMeminfoCmd, nodeDiskinfoCmd, nodeVncCmd)
	jobCmd.AddCommand(jobTraceCmd, jobMeminfoCmd)
	clusterCmd.AddCommand(qstatCmd, configCmd, matlabCmd, jobCmd, nodeCmd)

	rootCmd.AddCommand(clusterCmd)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Retrieve information about the HPC cluster or a job.",
	Long:  ``,
}

var qstatCmd = &cobra.Command{
	Use:   "qstat",
	Short: "Print job list in the memory of the Torque server.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("qstat command is triggerd.")
		c := trqhelper.TorqueHelperSrvClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintClusterQstat(xml); err != nil {
			log.Errorf("%+v\n", err)
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print Torque and Moab server configurations.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {

		if cmd.Flags().Changed("verbose") {
			log.SetLevel(log.DebugLevel)
		}

		c := trqhelper.TorqueHelperSrvClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintClusterConfig(); err != nil {
			log.Errorf("%+v\n", err)
		}
	},
}

var matlabCmd = &cobra.Command{
	Use:   "matlablic",
	Short: "Print a summary of the Matlab license usage.",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		stdout, stderr, ec, err := execCmd("lmstat", []string{"-a"})
		if err != nil {
			log.Fatalf("%s: exit code %d\n", err, ec)
		}
		if ec != 0 {
			log.Fatal(stderr.String())
		}

		rePkg := regexp.MustCompile(`^Users of (\S+):  \(Total of (\d+) licenses issued;  Total of (\d+) licenses in use\)$`)
		reUse := regexp.MustCompile(`^\s+(\S+) (\S+).*\((v[0-9]+)\).*, start (.*)$`)

		var lic matlabLicense
		var lics []matlabLicense
		for {
			line, err := stdout.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalln("fail parsing lmstat data")
			}

			if d := rePkg.FindAllStringSubmatch(line, -1); d != nil {

				log.Debugf("find license package: %s\n", line)

				// new license package found, put current lic into lics if the current lic is not nil
				if lic.Package != "" {
					lics = append(lics, lic)
				}

				// create a new matlabLicense with the parsed data
				n := d[0][1]
				t, _ := strconv.Atoi(d[0][2])
				lic = matlabLicense{Package: n, Total: t}

				continue
			}

			if d := reUse.FindAllStringSubmatch(line, -1); d != nil {
				log.Debugf("find package usage: %s\n", line)
				// new license usage found, parse it and add it to the license package's usage attribute.
				usage := matlabLicenseUsageInfo{User: d[0][1], Host: d[0][2], Version: d[0][3], Since: d[0][4]}
				lic.Usages = append(lic.Usages, usage)
				continue
			}
		}
		// print licenses
		for _, lic := range lics {
			fmt.Printf("\n%-24s %4d of %4d in use", lic.Package, len(lic.Usages), lic.Total)
		}
	},
}

// job related subcommands
var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Retrieve information about a cluster job.",
	Long:  ``,
}

var jobTraceCmd = &cobra.Command{
	Use:   "trace [JobID]",
	Short: "Print job's trace log available on the Torque server.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := trqhelper.TorqueHelperSrvClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintClusterTracejob(args[0]); err != nil {
			log.Errorf("fail get job trace info: %+v\n", err)
		}
	},
}

var jobMeminfoCmd = &cobra.Command{
	Use:   "meminfo [JobID]",
	Short: "Print memory usage of a running job.",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := trqhelper.TorqueHelperMomClient{
			SrvHost:     TorqueServerHost,
			SrvPort:     TorqueHelperPort,
			SrvCertFile: TorqueHelperCert,
		}
		if err := c.PrintJobMemoryInfo(args[0]); err != nil {
			log.Errorf("fail get job memory utilisation: %+v\n", err)
		}
	},
}

// node related subcommands
type nodeType uint

const (
	access nodeType = iota
	compute
)

var nodeTypeNames = map[nodeType]string{
	access:  "access",
	compute: "compute",
}

var nodeCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Retrieve information about cluster nodes.",
	Long:  ``,
	// ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	// Run: func(cmd *cobra.Command, args []string) {
	// 	// TODO: get nodes overview
	// },
}

var nodeMeminfoCmd = &cobra.Command{
	Use:       "memfree {access|compute}",
	Short:     "Print total and free memory on the cluster nodes.",
	Long:      ``,
	ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{nodeTypeNames[access], nodeTypeNames[compute]}
		}
		for _, n := range args {
			switch n {
			case nodeTypeNames[access]:
				g := dg.GangliaDataGetter{Dataset: dg.MemoryUsageAccessNode}
				g.GetPrint()
			case nodeTypeNames[compute]:
				g := dg.GangliaDataGetter{Dataset: dg.MemoryUsageComputeNode}
				g.GetPrint()
			}
		}
	},
}

var nodeDiskinfoCmd = &cobra.Command{
	Use:       "diskfree {access|compute}",
	Short:     "Print total and free disk space of the cluster nodes.",
	Long:      ``,
	ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{nodeTypeNames[access], nodeTypeNames[compute]}
		}
		for _, n := range args {
			switch n {
			case nodeTypeNames[access]:
				g := dg.GangliaDataGetter{Dataset: dg.DiskUsageAccessNode}
				g.GetPrint()
			case nodeTypeNames[compute]:
				g := dg.GangliaDataGetter{Dataset: dg.DiskUsageComputeNode}
				g.GetPrint()
			}
		}
	},
}

var nodeVncCmd = &cobra.Command{
	Use:   "vnc {hostname}",
	Short: "Print owner and display of running vnc servers on one of all of the cluster access nodes.",
	Long:  ``,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		// internal data structure to hold list of vncs by host
		type data struct {
			host string
			vncs []trqhelper.VNCServer
		}

		nodes := make(chan string, 4)
		vncservers := make(chan data)

		// worker group
		wg := new(sync.WaitGroup)
		nworker := 4
		wg.Add(nworker)

		// spin off two gRPC workers as go routines
		for i := 0; i < nworker; i++ {
			go func() {
				c := trqhelper.TorqueHelperAccClient{
					SrvPort:     TorqueHelperPort,
					SrvCertFile: TorqueHelperCert,
				}
				for h := range nodes {
					log.Debugf("work on %s", h)

					c.SrvHost = h
					servers, err := c.GetVNCServers()
					if err != nil {
						log.Errorf("%s: %s", c.SrvHost, err)
					}

					vncservers <- data{
						host: c.SrvHost,
						vncs: servers,
					}
				}

				log.Debugln("worker is about to leave")
				wg.Done()
			}()
		}

		// wait for all workers to finish
		go func() {
			wg.Wait()
			close(vncservers)
		}()

		// filling access node hosts
		go func() {
			// sort nodes
			sort.Strings(args)
			for _, n := range args {

				if !strings.HasSuffix(n, fmt.Sprintf(".%s", NetDomain)) {
					n = fmt.Sprintf("%s.%s", n, NetDomain)
				}

				log.Debugf("add node %s\n", n)
				nodes <- n
			}
			if len(args) == 0 {
				// TODO: append hostname of all of the access nodes.
				accs, err := dg.GetAccessNodes()
				// sort nodes
				sort.Strings(accs)
				if err != nil {
					log.Errorln(err)
				}
				for _, n := range accs {
					nodes <- n
				}
			}
			close(nodes)
		}()

		// reorganise internal data structure for sorting
		_hosts := []string{}
		_vncs := make(map[string][]trqhelper.VNCServer)
		for d := range vncservers {
			_hosts = append(_hosts, d.host)
			_vncs[d.host] = d.vncs
		}
		sort.Strings(_hosts)

		// simple display
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 0, 4, 0, ' ', 0)
		fmt.Fprintf(w, "\n%-10s\t%s\t", "Username", "VNC session")
		fmt.Fprintf(w, "\n%-10s\t%s\t", "--------", "-----------")
		for _, h := range _hosts {
			vncs := _vncs[h]
			sort.Slice(vncs, func(i, j int) bool {
				idi, _ := strconv.ParseUint(strings.Split(vncs[i].ID, ":")[1], 10, 32)
				idj, _ := strconv.ParseUint(strings.Split(vncs[j].ID, ":")[1], 10, 32)
				return idi < idj
			})
			for _, vnc := range vncs {
				fmt.Fprintf(w, "\n%-10s\t%s\t", vnc.Owner, vnc.ID)
			}
		}
		fmt.Fprintf(w, "\n")
		w.Flush()
	},
}

// execCmd executes a system call and returns stdout, stderr and exit code of the execution.
func execCmd(cmdName string, cmdArgs []string) (stdout, stderr bytes.Buffer, ec int32, err error) {
	// Execute command and catch the stdout and stderr as byte buffer.
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Env = os.Environ()
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	ec = 0
	if err = cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			ec = int32(ws.ExitStatus())
		} else {
			ec = 1
		}
	}
	return
}

// matlabLicense defines data structure of matlab license information and usage parsed from the
// `lmstat -a` command.
type matlabLicense struct {
	Package string
	Total   int
	Usages  []matlabLicenseUsageInfo
}

// matlabLicenseUsageInfo defines data structure of a matlab license that is in use.
type matlabLicenseUsageInfo struct {
	User    string
	Host    string
	Version string
	Since   string
}
