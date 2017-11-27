package acceptance

import (
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
	"testing"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type Options struct {
	ImmediateFail bool
	ProjectName   string
}

type AcceptanceFlow struct {
	options Options
	project project.APIProject
	client  hazelcast.IHazelcastInstance
	context project.Context
}

func NewFlow() AcceptanceFlow {
	flow := AcceptanceFlow{}
	flow.options.ImmediateFail = true
	flow.options.ProjectName = "hazelcast"
	return flow
}

func (flow AcceptanceFlow) Project(f string) AcceptanceFlow {
	flow.context = project.Context{
		ComposeFiles: []string{f},
		ProjectName:  flow.options.ProjectName,
	}

	project, err := docker.NewProject(&ctx.Context{
		Context: flow.context,
	}, nil)

	if err != nil && flow.options.ImmediateFail {
		log.Fatal(err)
	}

	flow.project = project
	return flow
}

func (flow AcceptanceFlow) Up() AcceptanceFlow {
	name := flow.options.ProjectName
	err := flow.project.Up(context.Background(), options.Up{}, name)
	if err != nil && flow.options.ImmediateFail {
		log.Fatal(err)
	}
	// todo improve wait on event
	time.Sleep(5 * time.Second)
	return flow
}

func (flow AcceptanceFlow) Scale() AcceptanceFlow {

	m := make(map[string]int)
	m[flow.options.ProjectName] = 3
	flow.project.Scale(context.Background(), 10000, m)

	// todo improve wait on event
	//time.Sleep(10 * time.Second)
	return flow
}

func (flow AcceptanceFlow) Down() AcceptanceFlow {
	err := flow.project.Down(context.Background(), options.Down{})
	if err != nil && flow.options.ImmediateFail {
		log.Fatal(err)
	}
	return flow
}

func (flow AcceptanceFlow) Client() AcceptanceFlow {
	config := hazelcast.NewHazelcastConfig()
	client, err := hazelcast.NewHazelcastClientWithConfig(config)
	if err != nil && flow.options.ImmediateFail {
		flow.Down()
		log.Fatal(err)
	}

	members := client.GetCluster().GetMemberList()
	log.Printf("Number of members : %v", len(members))
	flow.client = client
	return flow
}

func (flow AcceptanceFlow) TryMap(t *testing.T) AcceptanceFlow {
	mp, err := flow.client.GetMap(randSeq(42))
	if err != nil {
		flow.Down()
		log.Fatal(err)
	}

	key := randSeq(42)
	value := randSeq(42)

	mp.Put(key, value)
	actual, _ := mp.Get(key)

	assert.Equal(t, value, actual)
	return flow
}
