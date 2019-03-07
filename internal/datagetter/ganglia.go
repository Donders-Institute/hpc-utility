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
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
)

type gangliaDataset uint

const (
	gib float64 = 1024 * 1024 * 1024

	// MemoryUsageComputeNode is a predefined ganglia dataset of memory usage on the compute nodes.
	MemoryUsageComputeNode gangliaDataset = iota
	// DiskUsageComputeNode is a predefined ganglia dataset of disk usage on the compute nodes.
	DiskUsageComputeNode
	// MemoryUsageAccessNode is a predefined ganglia dataset of memory usage on the access nodes.
	MemoryUsageAccessNode
	// DiskUsageAccessNode is a predefined ganglia dataset of disk usage on the access nodes.
	DiskUsageAccessNode
)

// gangliaURLs defines ganglia data retrieval endpoints for different predefined datasets.
var gangliaURLs = map[gangliaDataset]string{
	MemoryUsageAccessNode:  "http://ganglia.dccn.nl/rawdata.php?c=Mentat%20Cluster&cols[]=mem_free.VAL&cols[]=mem_total.VAL&noheader",
	DiskUsageAccessNode:    "http://ganglia.dccn.nl/rawdata.php?c=Mentat%20Cluster&cols[]=disk_free.VAL&cols[]=disk_total.VAL&noheader",
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
	MemoryUsageComputeNode: scalerMib2gib,
	DiskUsageComputeNode:   scalerIdentical,
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
	// WriteHeader prints resource specific header row to the writer.
	WriteHeader(io.Writer) (int, error)
	// WriteData prints resource data row to the writer.
	WriteData(io.Writer, func(float64) float64) (int, error)
	// GetHostname returns hostname on which this resource is refers to.
	GetHostname() string
}

// gangliaMemdisk implements the gangliaResource interface for getting and reporting memory and disk resources.
type gangliaMemdisk struct {
	Host  string
	Free  float64
	Total float64
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

func (g gangliaMemdisk) WriteHeader(w io.Writer) (int, error) {
	n1, err := fmt.Fprintf(w, "\n %20s\t%10s\t%10s\t", "hostname", "free(GB)", "total(GB)")
	if err != nil {
		return n1, err
	}
	n2, err := fmt.Fprintf(w, "\n %20s\t%10s\t%10s\t", "--------", "--------", "---------")
	if err != nil {
		return n1 + n2, err
	}
	return n1 + n2, nil
}

func (g gangliaMemdisk) WriteData(w io.Writer, scaler func(float64) float64) (int, error) {
	return fmt.Fprintf(w, "\n %20s\t%10.1f\t%10.1f\t", g.Host, scaler(g.Free), scaler(g.Total))
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
		o := gangliaMemdisk{}
		err := g.parseGangliaRawdata(r, &o)
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

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', tabwriter.AlignRight)
	defer w.Flush()

	g.resources[0].WriteHeader(w)
	for _, r := range g.resources {
		r.WriteData(w, gangliaDataScaler[g.Dataset])
	}
	fmt.Fprintf(w, "\n")
}

// parseGangliaRawdata reads one record from the csv reader, and parse the data into the given
// data structure v.
func (g *GangliaDataGetter) parseGangliaRawdata(reader *csv.Reader, v interface{}) error {
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
		case "float64":
			fval, err := strconv.ParseFloat(record[i], 64)
			if err != nil {
				return err
			}
			f.SetFloat(fval)
		default:
			return fmt.Errorf("unsupported type: %s", f.Type().String())
		}
	}
	return nil
}
