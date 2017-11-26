package hz_go_it

import (
	"testing"
	"golang.org/x/net/context"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/hazelcast/go-client"
	"github.com/stretchr/testify/assert"
	"log"
	"time"
	"math/rand"
)

const (
	name = "hazelcast"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestSingleMemberConnection(t *testing.T) {
	project, err := docker.NewProject(&ctx.Context{
		Context: project.Context{
			ComposeFiles: []string{"./deployment.yaml"},
			ProjectName:  name,
		},
	}, nil)

	if err != nil {
		log.Fatal(err)
	}

	err = project.Up(context.Background(), options.Up{}, name)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	config := hazelcast.NewHazelcastConfig()
	client, err := hazelcast.NewHazelcastClientWithConfig(config)
	if err != nil {
		project.Down(context.Background(), options.Down{}, name)
		log.Fatal(err)
	}


	mp, err := client.GetMap(randSeq(42))
	if err != nil {
		project.Down(context.Background(), options.Down{}, name)
		log.Fatal(err)
	}

	key := randSeq(42)
	value := randSeq(42)

	mp.Put(key, value)
	actual, _ := mp.Get(key)

	assert.Equal(t, value, actual)

	err = project.Down(context.Background(), options.Down{})

	if err != nil {
		log.Fatal(err)
	}
}

