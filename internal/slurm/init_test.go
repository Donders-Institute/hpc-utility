package slurm

import (
	"strings"
	"testing"
)

var (
	nodeinfo = []string{
		`NodeName=dccn-c083 Arch=x86_64 CoresPerSocket=32
   CPUAlloc=2 CPUEfctv=63 CPUTot=64 CPULoad=3.00
   AvailableFeatures=(null)
   ActiveFeatures=(null)
   Gres=cpu:amd:1,gpu:nvidia_a100-sxm4-40gb:4(S:0-1)
   NodeAddr=dccn-c083 NodeHostName=dccn-c083 Version=22.05.10
   OS=Linux 4.18.0-553.8.1.el8_10.x86_64 #1 SMP Tue Jul 2 07:26:33 EDT 2024
   RealMemory=515578 AllocMem=106496 FreeMem=80395 Sockets=2 Boards=1
   CoreSpecCount=1 CPUSpecList=63 MemSpecLimit=4096
   State=MIXED ThreadsPerCore=1 TmpDisk=3604221 Weight=1 Owner=N/A MCS_label=N/A
   Partitions=gpu,batch
   BootTime=2024-11-14T17:10:50 SlurmdStartTime=2024-11-14T17:11:52
   LastBusyTime=2024-11-20T17:22:56
   CfgTRES=cpu=63,mem=515578M,billing=63,gres/gpu=4,gres/gpu:nvidia_a100-sxm4-40gb=4
   AllocTRES=cpu=2,mem=104G,gres/gpu=1,gres/gpu:nvidia_a100-sxm4-40gb=1
   CapWatts=n/a
   CurrentWatts=0 AveWatts=0
   ExtSensorsJoules=n/s ExtSensorsWatts=0 ExtSensorsTemp=n/s`,

		`NodeName=dccn-c084 Arch=x86_64 CoresPerSocket=32
   CPUAlloc=10 CPUEfctv=63 CPUTot=64 CPULoad=0.01
   AvailableFeatures=(null)
   ActiveFeatures=(null)
   Gres=cpu:amd:1,gpu:nvidia_a100-sxm4-40gb:4(S:0-1)
   NodeAddr=dccn-c084 NodeHostName=dccn-c084 Version=22.05.10
   OS=Linux 4.18.0-553.8.1.el8_10.x86_64 #1 SMP Tue Jul 2 07:26:33 EDT 2024
   RealMemory=515578 AllocMem=128000 FreeMem=375161 Sockets=2 Boards=1
   CoreSpecCount=1 CPUSpecList=63 MemSpecLimit=4096
   State=IDLE ThreadsPerCore=1 TmpDisk=3604221 Weight=1 Owner=N/A MCS_label=N/A
   Partitions=gpu,batch
   BootTime=2024-11-19T13:53:34 SlurmdStartTime=2024-11-19T13:54:41
   LastBusyTime=2024-11-20T15:16:28
   CfgTRES=cpu=63,mem=515578M,billing=63,gres/gpu=4,gres/gpu:nvidia_a100-sxm4-40gb=4
   AllocTRES=
   CapWatts=n/a
   CurrentWatts=0 AveWatts=0
   ExtSensorsJoules=n/s ExtSensorsWatts=0 ExtSensorsTemp=n/s`,
	}
)

func TestParseSingleNodeInfo(t *testing.T) {
	for _, info := range nodeinfo {
		info, err := parseSingleNodeInfo(info)
		if err != nil {
			t.Fatalf("%s\n", err)
		}

		t.Logf("node info: %+v\n", info)
	}
}

func TestParseMultipleNodeInfo(t *testing.T) {

	for _, node := range parseMultipleNodeInfo(strings.Join(nodeinfo, "\n")) {
		t.Logf("node info: %+v\n", node)
	}

}

func TestGetNodeInfo(t *testing.T) {
	nodes, err := GetNodeInfo("ALL")

	if err != nil {
		t.Fatalf("%s\n", err)
	}

	for _, node := range nodes {
		t.Logf("node info: %+v\n", node)
	}
}
