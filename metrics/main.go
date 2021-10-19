package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	var wg sync.WaitGroup
	var nodename = flag.String("n", "1704k8s00000000", "node to query")
	var cpuandmemory = flag.String("cpu", "?only_cpu_and_memory=true", "query for cpu and memory")
	var stats = flag.Int("l", 3, "list stats this many times")
	var start = flag.Int("s", 1, "start this many containers")
	var port = flag.String("p", "8001", "port")
	flag.Parse()

	fmt.Printf("Starting %d containers on node %s\n", *start, *nodename)
	c := listContainers(*nodename)
	fmt.Printf("pods at start: %d\n", len(c))

	if start != nil {
		fmt.Printf("starting %d containers\n", *start)
		for i := 0; i < *start; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runContainer("mcr.microsoft.com/windows/servercore/iis:windowsservercore-ltsc2019")
			}()
		}
	}

	fmt.Printf("run call to stats %d times\n", *stats)
	for i := 0; i < *stats; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := get(fmt.Sprintf("http://localhost:%s/api/v1/nodes/%s/proxy/stats/summary%s", *port, *nodename, *cpuandmemory))
			if err != nil {
				log.Printf("didn't find stats for container %v", err)
			}
		}()
	}

	wg.Wait()

	c = listContainers(*nodename)
	fmt.Printf("pods at end: %d\n", len(c))
}

// https://stackoverflow.com/a/45766707/697126
func timing(what string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", what, time.Since(start))
	}
}

func get(url string) (string, error) {
	defer timing(fmt.Sprintf("timing: %s", url))()
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func runContainer(image string) {
	s := rand.Int()
	c := fmt.Sprintf("kubectl run iis-%d -n default --image %s --image-pull-policy=Always --overrides '{ \"spec\": {\"nodeSelector\":{\"kubernetes.io/os\":\"windows\"}}}'", s, image)
	cmd := exec.Command("sh", "-c", c)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Container run err %v", err)
	}
	fmt.Printf("Container run successful: %q\n", out.String())
}

func listContainers(nodename string) []string {
	c := fmt.Sprintf("kubectl get pods -A --field-selector spec.nodeName=%s  --no-headers -o custom-columns=\":metadata.name\"", nodename)
	cmd := exec.Command("sh", "-c", c)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Container list err %v", err)
	}
	//fmt.Println(out.String())
	return strings.Split(out.String(), "\n")
}
