package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func main() {
	var wg sync.WaitGroup

	for _, m := range []int{500, 1000, 1500, 2000} {
		wg.Add(1)
		go Run(os.Args[1], m, &wg)
	}

	wg.Wait()
}

func Run(image string, millicores int, wg *sync.WaitGroup) {
	defer wg.Done()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatalln(err)
	}

	hostBinding := nat.PortBinding{
		HostIP: "0.0.0.0",
	}

	containerPort, err := nat.NewPort("tcp", "9000")
	if err != nil {
		log.Fatalln(err)
	}

	portBinding := nat.PortMap{
		containerPort: []nat.PortBinding{hostBinding},
	}

	platform := platforms.DefaultSpec()
	c, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: image,
			Tty:   true,
		},
		&container.HostConfig{
			PortBindings: portBinding,
			Resources: container.Resources{
				NanoCPUs: int64(millicores) * int64(math.Pow(10, 6)),
			},
			AutoRemove: true,
		}, nil, &platform, "")
	if err != nil {
		log.Fatalln(err)
	}

	statusChecker := http.Client{
		Timeout: 500 * time.Millisecond,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Millisecond,
			}).Dial,
		},
	}

	log.Println("container starting")
	start := time.Now()
	cli.ContainerStart(context.Background(), c.ID, types.ContainerStartOptions{})
	defer func() {
		timeout := 10 * time.Second
		cli.ContainerStop(context.Background(), c.ID, &timeout)
	}()

	inspect, err := cli.ContainerInspect(context.Background(), c.ID)
	hostmap := inspect.NetworkSettings.Ports[containerPort][0]
	url := fmt.Sprintf("http://%s:%s/status", hostmap.HostIP, hostmap.HostPort)
	log.Printf("%s: starting checker against '%s' with %dm CPU\n", inspect.Name, url, millicores)
	for {
		resp, _ := statusChecker.Get(url)
		if resp != nil && err == nil && resp.StatusCode == http.StatusOK {
			log.Printf("%s: success got, took %s\n", inspect.Name, time.Now().Sub(start))
			break
		}

		time.Sleep(1 * time.Second)
	}
}
