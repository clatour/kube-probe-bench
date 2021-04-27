package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func main() {
	durations := [5]time.Duration{}
	image := os.Args[1]
	port := os.Args[2]
	endpoint := os.Args[3]

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalln(err)
	}

	platform := platforms.DefaultSpec()

	log.Printf("ensuring image '%s' is pulled\n", image)
	rc, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{Platform: platform.Architecture})
	if err != nil {
		log.Fatalln(err)
	}

	// Discard whatever body the docker cli gives back
	io.Copy(io.Discard, rc)
	rc.Close()

	for i, m := range []int{500, 1000, 1500, 2000, 4000} {
		durations[i] = Run(cli, platform, image, port, endpoint, m)
	}

	for i := 0; i < len(durations)-1; i++ {
		diff := durations[i+1] - durations[i]
		log.Println(diff)
	}
}

func Run(cli *client.Client, platform v1.Platform, image, port, endpoint string, millicores int) time.Duration {
	hostBinding := nat.PortBinding{
		HostIP: "0.0.0.0",
	}

	containerPort, err := nat.NewPort("tcp", port)
	if err != nil {
		log.Fatalln(err)
	}

	portBinding := nat.PortMap{
		containerPort: []nat.PortBinding{hostBinding},
	}

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
	u, err := url.Parse(fmt.Sprintf("http://%s:%s%s", hostmap.HostIP, hostmap.HostPort, endpoint))
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%s (%dm cpu): starting probe against '%s'\n", inspect.Name, millicores, u)
	for {
		resp, _ := statusChecker.Get(u.String())
		if resp != nil && err == nil && resp.StatusCode == http.StatusOK {
			duration := time.Now().Sub(start)
			log.Printf("%s (%dm cpu): success after '%s'\n", inspect.Name, millicores, duration)
			return duration
		}

		time.Sleep(1 * time.Second)
	}
}
