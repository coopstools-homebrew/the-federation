package lib

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

type NodeHistogram struct {
	Data map[string][]topNode
	topNodeCaller func() ([]byte, error)
}

type topNode struct {
	cpu 	int64
	cpuPer	int
	mem 	int64
	memPer	int
}

func BuildNodeHistogram(topNodeCaller func() ([]byte, error)) NodeHistogram {
	if topNodeCaller == nil {
		topNodeCaller = func() ([]byte, error) {
			cmd := exec.Command("kubectl", "top", "nodes", "--use-protocol-buffers")
			return cmd.Output()
		}
	}
	return NodeHistogram{Data: map[string][]topNode{}, topNodeCaller: topNodeCaller}
}

func (histogram *NodeHistogram) SetupCron(interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Println("Running update")
				go histogram.UpdateStats()
			case <- quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (histagram *NodeHistogram) UpdateStats() {
	resp, err := histagram.topNodeCaller()
	if err != nil {
		fmt.Printf("%+v", errors.Wrap(err, "Trouble getting node info"))
		return
	}
	fmt.Printf("Function call: %s\n", resp)
	currentNodeData := parseRows(resp)
	fmt.Printf("Current output: %v\n", currentNodeData)
	for nodeName := range currentNodeData {
		if _, ok := histagram.Data[nodeName]; !ok {
			histagram.Data[nodeName] = []topNode{}
		}
	}
	for nodeName := range histagram.Data {
		if nodeMetrics, ok := currentNodeData[nodeName]; ok {
			histagram.Data[nodeName] = append(histagram.Data[nodeName], nodeMetrics)
		} else {
			histagram.Data[nodeName] = append(histagram.Data[nodeName], topNode{})
		}
		if len(histagram.Data[nodeName]) > 60 {
			histagram.Data[nodeName] = histagram.Data[nodeName][1:]
		}
	}
}

func parseRows(row []byte) map[string]topNode {
	pattern := regexp.MustCompile("([0-9a-zA-Z\\-]*)\\s+([0-9]+)m\\s+([0-9]+)%\\s+([0-9]+)([KMGT])i\\s+([0-9]+)%")
	rows := pattern.FindAllStringSubmatch(string(row), -1)
	nodes := make(map[string]topNode, len(rows))
	for _, row := range rows {
		cpu, _ 		:= strconv.ParseInt(row[2], 10, 64)
		cpuPer, _ 	:= strconv.Atoi(row[3])
		mem			:= convertMemoryToKilobytes(row[4], row[5])
		memPer, _ 	:= strconv.Atoi(row[6])
		nodes[row[1]] = topNode{
			cpu:    cpu,
			cpuPer: cpuPer,
			mem:    mem,
			memPer: memPer,
		}
	}
	return nodes
}

func convertMemoryToKilobytes(strValue string, multiplier string) int64 {
	value, _ := strconv.ParseInt(strValue, 10, 64)
	switch multiplier {
	case "M":
		return value * 1000
	case "G":
		return value * 1000 * 1000
	case "T":
		return value * 1000 * 1000 * 1000
	}
	return value
}
