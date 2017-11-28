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
	"math/rand"
	"testing"
	"github.com/hazelcast/go-client/config"
	"github.com/hazelcast/go-client/core"
	"github.com/lucasjones/reggen"
	"time"
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
	File          string
}

type Scaling struct {
	Count int
}

type AcceptanceFlow struct {
	options    Options
	project    project.APIProject
	client     hazelcast.IHazelcastInstance
	context    project.Context
	createdMap core.IMap
	config     *config.ClientConfig
}

func NewFlow() AcceptanceFlow {
	flow := AcceptanceFlow{}
	flow.options.ImmediateFail = true
	flow.options.ProjectName = "hazelcast"
	flow.options.File = "./deployment.yaml"
	return flow
}

func (flow AcceptanceFlow) Project() AcceptanceFlow {
	flow.context = project.Context{
		ComposeFiles: []string{flow.options.File},
		ProjectName:  flow.options.ProjectName,
	}

	project, err := docker.NewProject(&ctx.Context{
		Context: flow.context,
	}, nil)

	if err != nil && flow.options.ImmediateFail {
		panic(err)
	}

	flow.project = project
	return flow
}

func (flow AcceptanceFlow) Up() AcceptanceFlow {
	name := flow.options.ProjectName
	err := flow.project.Up(context.Background(), options.Up{}, name)
	if err != nil && flow.options.ImmediateFail {
		panic(err)
	}
	//// todo improve wait on event
	time.Sleep(5 * time.Second)
	return flow
}

func (flow AcceptanceFlow) Scale(options Scaling) AcceptanceFlow {

	m := make(map[string]int)
	m[flow.options.ProjectName] = options.Count
	flow.project.Scale(context.Background(), 10000, m)

	// todo improve wait on event
	//time.Sleep(10 * time.Second)
	return flow
}

func (flow AcceptanceFlow) Down() AcceptanceFlow {
	//todo check cluster size
	err := flow.project.Down(context.Background(), options.Down{})
	if err != nil && flow.options.ImmediateFail {
		panic(err)
	}
	return flow
}

func (flow AcceptanceFlow) DefaultClient() AcceptanceFlow {
	var clientConfig = hazelcast.NewHazelcastConfig()
	clientConfig.ClientNetworkConfig().SetConnectionAttemptLimit(5)
	clientConfig.ClientNetworkConfig().SetConnectionTimeout(2)
	return flow.Client(clientConfig)
}

func (flow AcceptanceFlow) Client(config *config.ClientConfig) AcceptanceFlow {
	client, err := hazelcast.NewHazelcastClientWithConfig(config)
	if err != nil && flow.options.ImmediateFail {
		flow.Down()
		panic(err)
	}

	members := client.GetCluster().GetMemberList()
	log.Printf("Number of members : %v", len(members))
	flow.client = client
	flow.config = config
	return flow
}

func (flow AcceptanceFlow) TryMap(t *testing.T, args ...int) AcceptanceFlow {
	map_name := randSeq(42)
	mp, err := flow.client.GetMap(map_name)
	if err != nil {
		flow.Down()
		t.Fatal(err)
	}

	size, _ := mp.Size()
	assert.Equal(t, size, int32(0))

	count, valueSize := countAndSize(args...)

	for i := 0; i < count; i++ {
		key := randSeq(42)
		value := randSeq(valueSize)

		mp.Put(key, value)
		actual, _ := mp.Get(key)

		assert.Equal(t, value, actual)
	}

	s, _ := mp.Size()
	assert.Equal(t, s, int32(count))
	mp.Clear()
	s, _ = mp.Size()
	assert.Equal(t, s, int32(0))

	flow.createdMap = mp
	return flow
}

func countAndSize(args ...int) (int, int) {
	if len(args) == 0 || args == nil {
		return 1, 1024
	}
	if len(args) == 1 {
		return args[0], int(1024)
	}
	if len(args) > 1 {
		return args[0], args[1]
	}
	return 1, 1
}

func (flow AcceptanceFlow) ExpectError(t *testing.T) AcceptanceFlow {
	assert.NotNil(t, flow.createdMap)

	key := randSeq(42)
	value := randSeq(42)

	_, err := flow.createdMap.Put(key, value)
	if err == nil {
		flow.Down()
		t.Fatal("Error expected!")
	} else {
		log.Printf("Error is %v", err)
	}

	return flow
}

func (flow AcceptanceFlow) ExpectConnection(t *testing.T, expected int) AcceptanceFlow {
	members := flow.client.GetCluster().GetMemberList()
	assert.Equal(t, expected, len(members))
	return flow
}

func (flow AcceptanceFlow) Predicate(t *testing.T) AcceptanceFlow {
	s, _ := flow.createdMap.Size(); if s > 0 {
		flow.createdMap.Clear()
	}

	const keyRegex = "[a-z]{42}"
	const valueRegex = "[0-9]{42}"

	keyGen, _ := reggen.NewGenerator(keyRegex)
	valueGen, err := reggen.NewGenerator(valueRegex)

	const size = 1024

	keys := make([]string, size)
	values := make([]string, size)

	for i := 0; i < size; i++ {
		keys[i] = keyGen.Generate(42)
		values[i] = valueGen.Generate(42)
		flow.createdMap.Put(keys[i], values[i])
	}

	s, _ = flow.createdMap.Size()
	assert.Equal(t, s, int32(size))

	entrySet, err := flow.createdMap.EntrySetWithPredicate(core.Regex("this", valueRegex))
	if err != nil {
		flow.Down()
		t.Fatalf("Predicate error %v", err)
	}
	assert.Equal(t, size, len(entrySet))

	actualValues, err := flow.createdMap.ValuesWithPredicate(core.Regex("this", valueRegex))
	if err != nil {
		t.Fatalf("Predicate error %v", err)
	}
	assert.Equal(t, size, len(actualValues))
	assert.Subsetf(t, values, actualValues, "Fails value check")

	//TODO below fails. inform go-client team
	keySet, err := flow.createdMap.KeySetWithPredicate(core.Regex("this", keyRegex))
	if err != nil {
		flow.Down()
		t.Fatalf("Predicate error %v", err)
	}
	assert.Equal(t, len(keySet), size)

	flow.createdMap.Clear()

	return flow
}

func (flow AcceptanceFlow) EntryProcessor(t *testing.T, expected string, processor *EntryProcessor) AcceptanceFlow {
	key, _ := reggen.Generate("^[a-z]", 42)
	val, _ := reggen.Generate("^[0-9]", 1024)

	_, err := flow.createdMap.Put(key, val); if err != nil {
		flow.Down()
		t.Fatal(err)
	}

	actual, err := flow.createdMap.ExecuteOnKey(key, processor); if err != nil {
		flow.Down()
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)

	flow.createdMap.Clear()
	return flow
}