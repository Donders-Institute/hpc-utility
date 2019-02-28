package cmd

import (
	"encoding/csv"
	pxml "encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var xml bool

const (
	gib int64 = 1024 * 1024 * 1024
)

// variable may be set at the build time to fix the default location for the TorqueHelper server certificate.
var defTorqueHelperCert string

// Ganglia URLs for retriving node resource information
var gangliaAccessNodeMeminfoURL = "http://ganglia.dccn.nl/rawdata.php?c=Mentat%20Cluster&cols[]=mem_total.VAL&cols[]=mem_free.VAL&noheader"
var gangliaAccessNodeDiskinfoURL = "http://ganglia.dccn.nl/rawdata.php?c=Mentat%20Cluster&cols[]=disk_total.VAL&cols[]=disk_free.VAL&noheader"

var gangliaComputeNodeMeminfoURL = "http://ganglia.dccn.nl/rawdata.php?c=Torque%20Cluster&cols[]=mem_total.VAL&cols[]=mem_free.VAL&noheader"
var gangliaComputeNodeDiskinfoURL = "http://ganglia.dccn.nl/rawdata.php?c=Torque%20Cluster&cols[]=disk_total.VAL&cols[]=disk_free.VAL&noheader"

func init() {

	qstatCmd.Flags().BoolVarP(&xml, "xml", "x", false, "XML output")

	clusterCmd.PersistentFlags().StringVarP(&TorqueServerHost, "server", "s", "torque.dccn.nl", "Torque server hostname")
	clusterCmd.PersistentFlags().IntVarP(&TorqueHelperPort, "port", "p", 60209, "Torque helper service port")
	clusterCmd.PersistentFlags().StringVarP(&TorqueHelperCert, "cert", "c", defTorqueHelperCert, "Torque helper service certificate")

	nodeCmd.AddCommand(nodeMeminfoCmd)
	jobCmd.AddCommand(jobTraceCmd, jobMeminfoCmd)
	clusterCmd.AddCommand(qstatCmd, configCmd, jobCmd, nodeCmd)

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
	Use:       "nodes",
	Short:     "Retrieve information about cluster nodes.",
	Long:      ``,
	ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: get nodes overview
	},
}

var nodeMeminfoCmd = &cobra.Command{
	Use:       "memfree",
	Short:     "Print total and free memory of the cluster's access nodes.",
	Long:      ``,
	ValidArgs: []string{nodeTypeNames[access], nodeTypeNames[compute]},
	Run: func(cmd *cobra.Command, args []string) {

		urls := map[string]string{
			nodeTypeNames[access]:  gangliaAccessNodeMeminfoURL,
			nodeTypeNames[compute]: gangliaComputeNodeMeminfoURL,
		}

		for _, n := range args {

			url, err := url.Parse(urls[n])
			if err != nil {
				log.Errorf("invalid URL: %s\n", url)
				continue
			}
			resources, err := getGangliaResources(url)
			if err != nil {
				log.Errorln(err)
				continue
			}
			printGangliaResources(resources, []string{"hostname", "free", "totla"})
		}
	},
}

// gangliaRawHTML is a data object for unmarshaling the HTML document retrieved from ganglia.
// The `Pre` attribute contains actual raw data of the resource information.
type gangliaRawHTML struct {
	XMLName pxml.Name `xml:"html"`
	Pre     string    `xml:"body>pre"`
}

// gangliaResource is a data object of a ganglia resource.
type gangliaResource struct {
	Host  string
	Free  int64
	Total int64
}

func printGangliaResources(resources []gangliaResource, headers []string) {
	// sort resources by hostname
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Host < resources[j].Host
	})

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 20, 0, '\t', 0)
	defer w.Flush()

	headers = headers[0:3]
	var bars []string
	for i := 0; i < len(headers); i++ {
		bars = append(bars, strings.Repeat("-", len(headers[i])))
	}

	fmt.Fprintf(w, "\n %s\t%s\t%s\t", headers[0], headers[1], headers[2])
	fmt.Fprintf(w, "\n %s\t%s\t%s\t", bars[0], bars[1], bars[2])
	for _, r := range resources {
		fmt.Fprintf(w, "\n %s\t%d\t%d\t", r.Host, r.Free*1024/gib, r.Total*1024/gib)
	}
}

// getGangliaResources retrieves raw data from a ganglia service endpoint, parses the raw data, and
// turns it into a slice of the gangliaResource data objects.
func getGangliaResources(url *url.URL) ([]gangliaResource, error) {

	var resources []gangliaResource

	// make HTTP GET call to the ganglia endpoint.
	c := &http.Client{}
	req, _ := http.NewRequest("GET", url.String(), nil)
	resp, err := c.Do(req)
	if err != nil {
		return resources, err
	}
	if resp.StatusCode != 200 {
		return resources, fmt.Errorf("fail get ganglia data: %s (HTTP CODE %d)", url.String(), resp.StatusCode)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// the tabular data is enclosed by <pre></pre> tag in the returned HTML.
	data := gangliaRawHTML{}
	if err := pxml.Unmarshal(body, &data); err != nil {
		return resources, fmt.Errorf("fail get ganglia data: %s", err)
	}

	log.Debugf("ganglia tabular data: %s\n", data.Pre)

	// parse the tabular data
	r := csv.NewReader(strings.NewReader(data.Pre))
	r.Comma = '\t'
	r.TrailingComma = true
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	for {
		o := gangliaResource{}
		err := parseGangliaRawdata(r, &o)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Debugf("%+v\n", err)
			continue
		}
		resources = append(resources, o)
	}
	return resources, nil
}

// parseGangliaRawdata reads one record from the csv reader, and parse the data into the given
// data structure v.
func parseGangliaRawdata(reader *csv.Reader, v interface{}) error {
	record, err := reader.Read()
	if err != nil {
		return err
	}

	s := reflect.ValueOf(v).Elem()
	if len(record) < s.NumField() {
		return fmt.Errorf(fmt.Sprintf("expect %d field, got %d: %s", s.NumField(), len(record), record))
	}

	// reduce record to the expected amount of fields.
	record = record[0:s.NumField()]

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		switch f.Type().String() {
		case "string":
			f.SetString(record[i])
		case "int64":
			ival, err := strconv.ParseInt(record[i], 10, 0)
			if err != nil {
				return err
			}
			f.SetInt(ival)
		default:
			return fmt.Errorf("unsupported type: %s", f.Type().String())
		}
	}
	return nil
}
