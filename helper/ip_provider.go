package helper

import (
	"github.com/fsouza/go-dockerclient"
	"log"
)

func GetMemberIp(id string) []string{
	client, _ := docker.NewClientFromEnv()
	container, _ := client.InspectContainer(id)

	ips := make([]string, len(container.NetworkSettings.Networks))
	idx := 0
	for nt, network := range container.NetworkSettings.Networks {
		log.Print("Address " + network.IPAddress)
		log.Print("Type " + nt)
		ips[idx] = network.IPAddress
	}
	return ips
}
