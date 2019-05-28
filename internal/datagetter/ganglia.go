package datagetter

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

type gangliaDataset uint

const (
	gib float64 = 1024 * 1024 * 1024

	// MemoryUsageComputeNode is a predefined ganglia dataset of memory usage on the compute nodes.
	MemoryUsageComputeNode gangliaDataset = iota
	// DiskUsageComputeNode is a predefined ganglia dataset of disk usage on the compute nodes.
	DiskUsageComputeNode
	// InfoComputeNode is a predefined ganglia dataset of resource usage, system load on the compute nodes.
	InfoComputeNode
	// MemoryUsageAccessNode is a predefined ganglia dataset of memory usage on the access nodes.
	MemoryUsageAccessNode
	// DiskUsageAccessNode is a predefined ganglia dataset of disk usage on the access nodes.
	DiskUsageAccessNode
	// InfoAccessNode is a predefined ganglia dataset of resource usage, system load on the access nodes.
	InfoAccessNode
)

// gangliaURLs defines ganglia data retrieval endpoints for different predefined datasets.
var gangliaURLs = map[gangliaDataset]string{
	InfoAccessNode:         "http://ganglia.dccn.nl/rawdata.php?c=Mentat%20Cluster&cols[]=load_one.VAL&cols[]=load_five.VAL&cols[]=mem_free.VAL&cols[]=mem_total.VAL&cols[]=disk_free.VAL&cols[]=disk_total.VAL&noheader",
	MemoryUsageAccessNode:  "http://ganglia.dccn.nl/rawdata.php?c=Mentat%20Cluster&cols[]=mem_free.VAL&cols[]=mem_total.VAL&noheader",
	DiskUsageAccessNode:    "http://ganglia.dccn.nl/rawdata.php?c=Mentat%20Cluster&cols[]=disk_free.VAL&cols[]=disk_total.VAL&noheader",
	InfoComputeNode:        "http://ganglia.dccn.nl/rawdata.php?c=Torque%20Cluster&cols[]=load_one.VAL&cols[]=load_five.VAL&cols[]=mem_free.VAL&cols[]=mem_total.VAL&cols[]=disk_free.VAL&cols[]=disk_total.VAL&noheader",
	MemoryUsageComputeNode: "http://ganglia.dccn.nl/rawdata.php?c=Torque%20Cluster&cols[]=mem_free.VAL&cols[]=mem_total.VAL&noheader",
	DiskUsageComputeNode:   "http://ganglia.dccn.nl/rawdata.php?c=Torque%20Cluster&cols[]=disk_free.VAL&cols[]=disk_total.VAL&noheader",
}

// scalerMib2gib is a data unit scaler converting MiB to GiB.
func scalerMib2gib(x float64) float64 {
	return x * 1024 / gib
}

// scalerIdentical is a data unit scaler converting
func scalerIdentical(x float64) float64 {
	return x
}

// GetAccessNodes return the hostname of torque cluster's access nodes.
func GetAccessNodes() ([]string, error) {
	g := GangliaDataGetter{Dataset: MemoryUsageAccessNode}
	if err := g.Get(); err != nil {
		return nil, err
	}

	nodes := []string{}
	for _, r := range g.resources {
		nodes = append(nodes, r.GetHostname())
	}
	return nodes, nil
}

// GetComputeNodes return the hostname of torque cluster's compute nodes.
func GetComputeNodes() ([]string, error) {
	g := GangliaDataGetter{Dataset: MemoryUsageComputeNode}
	if err := g.Get(); err != nil {
		return nil, err
	}

	nodes := []string{}
	for _, r := range g.resources {
		nodes = append(nodes, r.GetHostname())
	}
	return nodes, nil
}

// gangliaDataScaler defines data scaler for different predefined datasets.
var gangliaDataScaler = map[gangliaDataset]func(float64) float64{
	MemoryUsageAccessNode:  scalerMib2gib,
	DiskUsageAccessNode:    scalerIdentical,
	InfoAccessNode:         scalerIdentical,
	MemoryUsageComputeNode: scalerMib2gib,
	DiskUsageComputeNode:   scalerIdentical,
	InfoComputeNode:        scalerIdentical,
}

// gangliaRawHTML is a data object for unmarshaling the HTML document retrieved from ganglia.
// The `Pre` attribute contains actual raw data of the resource information.
type gangliaRawHTML struct {
	XMLName xml.Name `xml:"html"`
	Pre     string   `xml:"body>pre"`
}

// gangliaResource defines the generic interface of a ganglia resource object.
type gangliaResource interface {
	// Less checks whether this resource object is objectically smaller than the other resource object.
	Less(*gangliaResource) bool
	// Parse reads one CSV row from the csv.Reader and converts it into a new gangliaResource.
	Parse(*csv.Reader) (gangliaResource, error)
	// GetTableWriterHeaders returns a slice of strings to be used by the tablewriter as tabular header.
	GetTableWriterHeaders() []string
	// GetTableWriterRowData() returns a slice of strings representing the resource as a data row in the tablewriter.
	GetTableWriterRowData(func(float64) float64) []string
	// GetHostname returns hostname on which this resource is refers to.
	GetHostname() string
}

// gangliaSysinfo implements the gangliaResource interface for getting and reporting system load, memory and disk resources.
type gangliaSysinfo struct {
	Host      string
	Load1     float64
	Load5     float64
	MemFree   float64
	MemTotal  float64
	DiskFree  float64
	DiskTotal float64
}

func (g gangliaSysinfo) Parse(reader *csv.Reader) (gangliaResource, error) {

	record, err := reader.Read()
	if err != nil {
		return nil, err
	}

	if len(record) < 7 {
		return nil, fmt.Errorf("invalid data record: %+v", record)
	}

	r := gangliaSysinfo{}

	r.Host = record[0]
	if r.Load1, err = strconv.ParseFloat(record[1], 64); err != nil {
		return nil, err
	}
	if r.Load5, err = strconv.ParseFloat(record[2], 64); err != nil {
		return nil, err
	}
	if r.MemFree, err = strconv.ParseFloat(record[3], 64); err != nil {
		return nil, err
	}
	if r.MemTotal, err = strconv.ParseFloat(record[4], 64); err != nil {
		return nil, err
	}
	if r.DiskFree, err = strconv.ParseFloat(record[5], 64); err != nil {
		return nil, err
	}
	if r.DiskTotal, err = strconv.ParseFloat(record[6], 64); err != nil {
		return nil, err
	}
	return r, nil
}

func (g gangliaSysinfo) GetHostname() string {
	return g.Host
}

func (g gangliaSysinfo) Less(other *gangliaResource) bool {
	o := reflect.ValueOf(other).Elem().Elem()
	return g.Host < o.FieldByName("Host").String()
}

func (g gangliaSysinfo) GetTableWriterHeaders() []string {
	return []string{"hostname", "load1", "load5", "mem free(GB)", "mem total(GB)", "disk free(GB)", "disk total(GB)"}
}

func (g gangliaSysinfo) GetTableWriterRowData(scaler func(float64) float64) []string {
	// in this function, the scaler is ignored as we know what to do with the scaling of the values.
	return []string{
		g.Host,
		fmt.Sprintf("%.1f", g.Load1),
		fmt.Sprintf("%.1f", g.Load5),
		fmt.Sprintf("%.1f", scalerMib2gib(g.MemFree)),
		fmt.Sprintf("%.1f", scalerMib2gib(g.MemTotal)),
		fmt.Sprintf("%.1f", g.DiskFree),
		fmt.Sprintf("%.1f", g.DiskTotal),
	}
}

// gangliaMemdisk implements the gangliaResource interface for getting and reporting memory and disk resources.
type gangliaMemdisk struct {
	Host  string
	Free  float64
	Total float64
}

func (g gangliaMemdisk) Parse(reader *csv.Reader) (gangliaResource, error) {

	record, err := reader.Read()
	if err != nil {
		return nil, err
	}

	if len(record) < 3 {
		return nil, fmt.Errorf("invalid data record: %+v", record)
	}

	r := gangliaMemdisk{}

	r.Host = record[0]
	if r.Free, err = strconv.ParseFloat(record[1], 64); err != nil {
		return nil, err
	}
	if r.Total, err = strconv.ParseFloat(record[2], 64); err != nil {
		return nil, err
	}
	return r, nil
}

func (g gangliaMemdisk) GetHostname() string {
	return g.Host
}

func (g gangliaMemdisk) Less(other *gangliaResource) bool {

	// t := reflect.TypeOf(other).Elem()
	// if !t.ConvertibleTo(reflect.TypeOf(gangliaMemdisk{})) {
	// 	fmt.Printf("type: %s\n", t.String())
	// 	panic(42)
	// }

	// TODO: check type of other!!
	o := reflect.ValueOf(other).Elem().Elem()
	return g.Host < o.FieldByName("Host").String()
}

func (g gangliaMemdisk) GetTableWriterHeaders() []string {
	return []string{"hostname", "free(GB)", "total(GB)"}
}

func (g gangliaMemdisk) GetTableWriterRowData(scaler func(float64) float64) []string {
	return []string{
		g.Host,
		fmt.Sprintf("%.1f", scaler(g.Free)),
		fmt.Sprintf("%.1f", scaler(g.Total)),
	}
}

// GangliaDataGetter provides interfaces to retrieve data from the Ganglia website, parse it
// and return relevant data objects.
type GangliaDataGetter struct {
	Dataset   gangliaDataset
	resources []gangliaResource
}

// GetPrint retrieves ganglia resource data and print the data to the stdout in a tabular format.
func (g *GangliaDataGetter) GetPrint() error {
	err := g.Get()
	if err != nil {
		return err
	}
	g.print()
	return nil
}

// newResource returns a gangaliaResource struct depending on the type of gangliaDataset.
func (g *GangliaDataGetter) newResource() gangliaResource {
	switch g.Dataset {
	case InfoAccessNode:
		return gangliaSysinfo{}
	case InfoComputeNode:
		return gangliaSysinfo{}
	default:
		return gangliaMemdisk{}
	}
}

// Get retrieves raw data from a ganglia service endpoint, parses the raw data, and
// turns it into a slice of the gangliaResource data objects.
func (g *GangliaDataGetter) Get() error {

	// make HTTP GET call to the ganglia endpoint.
	c := &http.Client{}
	req, _ := http.NewRequest("GET", gangliaURLs[g.Dataset], nil)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("fail get ganglia data: %s (HTTP CODE %d)", gangliaURLs[g.Dataset], resp.StatusCode)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// the tabular data is enclosed by <pre></pre> tag in the returned HTML.
	data := gangliaRawHTML{}
	if err := xml.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("fail get ganglia data: %s", err)
	}

	log.Debugf("ganglia tabular data: %s\n", data.Pre)

	// parse the tabular data
	r := csv.NewReader(strings.NewReader(data.Pre))
	r.Comma = '\t'
	r.TrailingComma = true
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	for {
		o, err := g.newResource().Parse(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Debugf("%+v\n", err)
			continue
		}
		g.resources = append(g.resources, o)
	}
	return nil
}

// Print writes retrieved ganglia resource data in a tabular format.
func (g *GangliaDataGetter) print() {

	// sort resources by hostname
	sort.Slice(g.resources, func(i, j int) bool {
		o := g.resources[j]
		return g.resources[i].Less(&o)
	})

	// create and write the tabular output using tablewriter
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(g.resources[0].GetTableWriterHeaders())
	for _, r := range g.resources {
		table.Append(r.GetTableWriterRowData(gangliaDataScaler[g.Dataset]))
	}
	table.Render()
}
