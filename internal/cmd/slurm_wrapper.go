package cmd

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	trqhelper "github.com/Donders-Institute/hpc-torque-helper/pkg/client"
	log "github.com/sirupsen/logrus"
)

// parseSingleNodeInfo converts the output of `scontrol show node <node>` into
// the `trqhelper.NodeResourceStatus` data structure.
//
// The expected `out` looks like the one below:
//
// ```
// NodeName=dccn-c084 Arch=x86_64 CoresPerSocket=32
//
//	CPUAlloc=0 CPUEfctv=63 CPUTot=64 CPULoad=0.01
//	AvailableFeatures=(null)
//	ActiveFeatures=(null)
//	Gres=cpu:amd:1,gpu:nvidia_a100-sxm4-40gb:4(S:0-1)
//	NodeAddr=dccn-c084 NodeHostName=dccn-c084 Version=22.05.10
//	OS=Linux 4.18.0-553.8.1.el8_10.x86_64 #1 SMP Tue Jul 2 07:26:33 EDT 2024
//	RealMemory=515578 AllocMem=0 FreeMem=375161 Sockets=2 Boards=1
//	CoreSpecCount=1 CPUSpecList=63 MemSpecLimit=4096
//	State=IDLE ThreadsPerCore=1 TmpDisk=3604221 Weight=1 Owner=N/A MCS_label=N/A
//	Partitions=gpu,batch
//	BootTime=2024-11-19T13:53:34 SlurmdStartTime=2024-11-19T13:54:41
//	LastBusyTime=2024-11-20T15:16:28
//	CfgTRES=cpu=63,mem=515578M,billing=63,gres/gpu=4,gres/gpu:nvidia_a100-sxm4-40gb=4
//	AllocTRES=
//	CapWatts=n/a
//	CurrentWatts=0 AveWatts=0
//	ExtSensorsJoules=n/s ExtSensorsWatts=0 ExtSensorsTemp=n/s
//
// ```
func parseSingleNodeInfo(out string) (trqhelper.NodeResourceStatus, error) {

	info := trqhelper.NodeResourceStatus{}

	cpuAllocated := 0
	memAllocated := 0
	gpuAllocated := 0

	var err error

	for _, field := range strings.Fields(out) {
		if keyValue := strings.SplitN(field, "=", 2); len(keyValue) == 2 {
			switch keyValue[0] {
			case "NodeName":
				info.ID = keyValue[1]
			case "CPUEfctv":
				nproc, err := strconv.Atoi(keyValue[1])
				if err != nil {
					log.Errorf("invalid number of total CPUs: %s", err)
					return info, err
				}
				info.TotalProcs = nproc
			case "CPUAlloc":
				if cpuAllocated, err = strconv.Atoi(keyValue[1]); err != nil {
					log.Errorf("invalid number of allocated CPUs: %s", err)
					return info, err
				}
			case "RealMemory":
				memMB, err := strconv.Atoi(keyValue[1])
				if err != nil {
					log.Errorf("invalid number of total memory: %s", err)
					return info, err
				}
				info.TotalMemGB = memMB / 1024
			case "AllocMem":
				if memAllocated, err = strconv.Atoi(keyValue[1]); err != nil {
					log.Errorf("invalid number of total memory: %s", err)
					return info, err
				}
			case "Partitions":
				info.Features = append(info.Features, strings.Split(keyValue[1], ",")...)

			case "State":
				info.State = strings.Split(keyValue[1], "+")[0]

			case "Gres":
				reCpu := regexp.MustCompile(`cpu:(amd|intel):[0-9]+`)

				cpuinfo := reCpu.FindStringSubmatch(keyValue[1])

				if len(cpuinfo) < 2 {
					return info, fmt.Errorf("unexpected CPU information: %s", keyValue[1])
				}
				switch cpuinfo[1] {
				case "amd":
					info.IsAMD = true
					info.IsIntel = false
				case "intel":
					info.IsIntel = true
					info.IsAMD = false
				}

			case "CfgTRES":
				reGpu := regexp.MustCompile(`gres/gpu=([0-9]+)`)
				gpuinfo := reGpu.FindStringSubmatch(keyValue[1])
				if len(gpuinfo) == 2 {
					ngpus, err := strconv.Atoi(gpuinfo[1])
					if err != nil {
						log.Errorf("invalid number of total GPUs: %s", err)
						return info, err
					}
					info.TotalGPUS = ngpus
				}

			case "AllocTRES":
				reGpu := regexp.MustCompile(`gres/gpu=([0-9]+)`)
				gpuinfo := reGpu.FindStringSubmatch(keyValue[1])
				if len(gpuinfo) == 2 {
					if gpuAllocated, err = strconv.Atoi(gpuinfo[1]); err != nil {
						log.Errorf("invalid number of allocated GPUs: %s", err)
						return info, err
					}
				}

			case "TmpDisk":
				diskMB, err := strconv.Atoi(keyValue[1])
				if err != nil {
					log.Errorf("invalid number of tmp disk size: %s", err)
					return info, err
				}
				info.TotalDiskGB = diskMB / 1024

			default:
				// do nothing
			}
		}
	}

	if info.ID == "" {
		return info, fmt.Errorf("invalid node: ID is empty")
	}

	info.AvailProcs = info.TotalProcs - cpuAllocated
	info.AvailGPUS = info.TotalGPUS - gpuAllocated
	info.AvailMemGB = info.TotalMemGB - (memAllocated / 1024)

	return info, nil

}

func parseMultipleNodeInfo(out string) []trqhelper.NodeResourceStatus {

	nodes := make([]trqhelper.NodeResourceStatus, 0)

	for _, nodeinfo := range strings.Split(out, "NodeName=") {

		if nodeinfo == "" {
			continue
		}

		node, err := parseSingleNodeInfo(fmt.Sprintf("NodeName=%s", nodeinfo))
		if err != nil {
			log.Errorf("%s", err)
			continue
		}
		nodes = append(nodes, node)
	}

	return nodes
}

// GetNodeInfo makes a system call `scontrol show node` and parse the
// output into array of `trqhelper.NodeResourceStatus`.
//
// If the given argument `id` is a empty string `""“ or `"ALL"`, it will
// get information of all Slurm nodes.
func GetNodeInfo(id string) ([]trqhelper.NodeResourceStatus, error) {

	args := []string{"show", "node"}

	if id != "" && id != "ALL" {
		args = append(args, id)
	}

	stdout, stderr, ec, err := execCmd("scontrol", args)

	if err != nil {
		log.Fatalf("%s: exit code %d\n", err, ec)
	}
	if ec != 0 {
		log.Fatal(stderr.String())
	}

	nodeInfo := ""
	nodes := make([]trqhelper.NodeResourceStatus, 0)

	for {
		line, err := stdout.ReadString('\n')

		if err == io.EOF {
			if nodeInfo != "" {
				node, err := parseSingleNodeInfo(nodeInfo)

				if err != nil {
					log.Errorf("%s", err)
				}
				nodes = append(nodes, node)
			}
			break
		}

		// non-EOF error
		if err != nil {
			log.Fatalln("fail reading scontrol data")
		}

		if strings.HasPrefix(line, "NodeName=") && nodeInfo != "" {
			node, err := parseSingleNodeInfo(nodeInfo)
			if err != nil {
				log.Errorf("%s", err)
			}
			nodes = append(nodes, node)

			// reset nodeInfo
			nodeInfo = line
		} else {
			nodeInfo = fmt.Sprintf("%s%s", nodeInfo, line)
		}
	}

	return nodes, nil
}
