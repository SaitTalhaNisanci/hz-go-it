package acceptance

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"golang.org/x/net/context"
	"github.com/docker/docker/client"
	"strings"
	"log"
	"os/exec"
)

func find_container_ip(id string) string {
	apiClient, _ := client.NewEnvClient()
	var opts = types.NetworkListOptions{
		Filters: filters.NewArgs(),
	}
	resources, _ := apiClient.NetworkList(context.Background(), opts)

	var ip = "";
	for _, resource := range resources {
		if resource.Name == "go-it" {
			container := resource.Containers[id]
			if strings.Contains(container.Name, "hazelcast") {
				ip = strings.Split(container.IPv4Address, "/")[0]
			}
		}
	}

	return ip
}

func wait_for_port(ip string) {
	commandStr := "./wait.sh " + ip + ":" + "5701" + " -t 10"
	cmd := exec.Command("/bin/sh", "-c", commandStr)
	_, err := cmd.Output()
	if err != nil {
		log.Print("Error on wait for port " + err.Error())
	}
}
