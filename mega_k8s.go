package main

import (
	"errors"
	"strconv"
	"strings"
)

//map types
type MP map[string]interface{}
type MPArray []MP

type JSONString map[string]string

const APIFormat string = "%s:%s%s"

// Header format : Header-X: data\n Header-O: data0
const APIAuthHeaders string = ""

func RequestPods() (list PodMetricList, err error) {
	jsonData := GetMetricData()
	err = ParsePodMetricList(&list, jsonData)
	return
}
func (list PodMetricList) CheckContainers(pod PodConfig, contacts []Contact, mailcfg MailSettings, smsinfo TwilioInfo) {
	if pod.Watching {
		metrics, err := list.GetPodMetrics(pod.Name)
		if err != nil {
			NotifyPodContacts(pod, contacts, mailcfg, smsinfo, " no data found.")
		}

		for _, metric := range metrics {
			cpuUsage := metric.GetCPUUsage()
			if cpuUsage > pod.MaxCPU {
				NotifyPodContacts(pod, contacts, mailcfg, smsinfo, " is using more than its allocated CPU resources.")
			}
			memoryUsage := metric.GetMemoryUsage()
			if memoryUsage > pod.MaxMemory {
				NotifyPodContacts(pod, contacts, mailcfg, smsinfo, " is using more than its allocated memory resources.")
			}
		}

	}
}

func (list PodMetricList) GetPodMetrics(name string) (result MPArray, err error) {
	for _, pods := range list.Items {

		for _, podMetric := range pods.Containers {
			if podMetric["name"].(string) == name {
				result = append(result, podMetric["usage"].(map[string]interface{}))
			}
		}

	}
	LengthRes := len(result)
	if LengthRes == 0 {
		err = errors.New("No metrics found with container name.")
	}
	return
}

func (usage MP) GetCPUUsage() (cputime int64) {
	cpustr := usage["cpu"].(string)
	if strings.Contains(cpustr, "m") && strings.Contains(cpustr, "s") {
		mspl := strings.Split(cpustr, "m")
		sspl := strings.Split(mspl[1], "s")
		minutes, _ := strconv.ParseInt(mspl[0], 10, 64)
		seconds, _ := strconv.ParseInt(sspl[0], 10, 64)
		cputime = int64(float64(minutes)/60.0) + seconds

	} else if strings.Contains(cpustr, "m") {
		mspl := strings.Split(cpustr, "m")
		minutes, _ := strconv.ParseInt(mspl[0], 10, 64)
		cputime = int64(float64(minutes) / 60.0)

	} else if strings.Contains(cpustr, "s") {
		mspl := strings.Split(cpustr, "s")
		cputime, _ = strconv.ParseInt(mspl[0], 10, 64)

	}
	return
}

func (usage MP) GetMemoryUsage() (i int64) {
	memorystr := usage["memory"].(string)
	var bspl []string
	if strings.Contains(memorystr, "Mi") {
		bspl = strings.Split(memorystr, "Mi")
		i, _ = strconv.ParseInt(bspl[0], 10, 64)

	} else if strings.Contains(memorystr, "Ki") {
		bspl = strings.Split(memorystr, "Ki")
		i, _ = strconv.ParseInt(bspl[0], 10, 64)
		i = int64(float64(i) / 1000.0)
	} else {
		bspl = strings.Split(memorystr, "Gi")
		i, _ = strconv.ParseInt(bspl[0], 10, 64)
		i = int64(float64(i) * 1000.0)
	}

	return
}
