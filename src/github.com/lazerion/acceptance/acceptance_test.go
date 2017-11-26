package hz_go_it

import (
	"testing"
	"golang.org/x/net/context"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/hazelcast/go-client"
	"log"
)

func TestClusterConnection(t *testing.T){
	project, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{"./deployment.yaml"},
			ProjectName:  "hazelcast-cluster",
		},
	}, nil)

	if err != nil {
		log.Fatal(err)
	}

	err = project.Up(context.Background(), options.Up{})

	if err != nil {
		log.Fatal(err)
	}

	config := hazelcast.NewHazelcastConfig()
	client, _ := hazelcast.NewHazelcastClientWithConfig(config)

	log.Printf("config %v", config)
	log.Printf("client %v", client)

	mp, _ := client.GetMap("myMap")

	mp.Put("test", "test")
	val, _ := mp.Get("test")

	log.Printf("%v", val)

	err = project.Down(context.Background(), options.Down{})

	if err != nil {
		log.Fatal(err)
	}
}

