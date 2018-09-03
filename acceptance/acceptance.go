package acceptance

import (
	"golang.org/x/net/context"
	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/docker/ctx"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/project/options"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"github.com/hazelcast/hazelcast-go-client/config"
	"github.com/hazelcast/hazelcast-go-client/core"
	"github.com/hazelcast/hazelcast-go-client/core/predicate"
	"github.com/lucasjones/reggen"
	"time"
	"sync"
	"github.com/montanaflynn/stats"
)

type Options struct {
	ImmediateFail bool
	ProjectName   string
	File          string
	Store         bool
}

type Scaling struct {
	Count int
}

type Store struct {
	entry   map[string]string
	mapName string
}

type AcceptanceFlow struct {
	options    Options
	project    project.APIProject
	client     hazelcast.Instance
	context    project.Context
	createdMap core.Map
	config     *config.Config
	samples    []float64
	memberIp   []string
	store      Store
}

func NewFlow() AcceptanceFlow {
	flow := AcceptanceFlow{}
	flow.options.ImmediateFail = true
	flow.options.ProjectName = "hazelcast-go-it"
	flow.options.File = "./deployment.yaml"
	flow.options.Store = false
	flow.store = Store{entry:make(map[string]string), mapName:""}
	return flow
}

func (flow AcceptanceFlow) Project() AcceptanceFlow {
	flow.context = project.Context{
		ComposeFiles: []string{flow.options.File},
		ProjectName:  flow.options.ProjectName,
	}

	prj, err := docker.NewProject(&ctx.Context{
		Context: flow.context,
	}, nil)

	if err != nil && flow.options.ImmediateFail {
		panic(err)
	}

	flow.project = prj
	return flow
}

func (flow AcceptanceFlow) Up() AcceptanceFlow {
	name := flow.options.ProjectName

	err := flow.project.Up(context.Background(), options.Up{}, name)
	if err != nil && flow.options.ImmediateFail {
		panic(err)
	}

	containers, err := flow.project.Containers(context.Background(), project.Filter{
		State: project.Running,
	}, flow.options.ProjectName)

	if err != nil && flow.options.ImmediateFail {
		panic(err)
	}

	ip := find_container_ip(containers[0])
	log.Printf("IP %v", ip)
	flow.memberIp = []string{ip}
	wait_for_port(ip)

	return flow
}

func (flow AcceptanceFlow) Scale(options Scaling) AcceptanceFlow {

	m := make(map[string]int)
	m[flow.options.ProjectName] = options.Count
	flow.project.Scale(context.Background(), int(10 * time.Second), m)

	containers, _ := flow.project.Containers(context.Background(), project.Filter{
		State: project.Running,
	}, flow.options.ProjectName)

	for _, container := range containers {
		ip := find_container_ip(container)
		wait_for_port(ip)
	}

	return flow
}

func (flow AcceptanceFlow) ClusterSize(t *testing.T, expected int) AcceptanceFlow {

	const tryCount = 60
	const tryTimeout = 2 * time.Second

	var actual = 0
	for idx := 0; idx < tryCount; idx++ {
		actual = len(flow.client.Cluster().GetMembers())
		if actual != expected {
			time.Sleep(tryTimeout)
			log.Printf("Cluster size is not met, retrying actual %d, expected %d", actual, expected)
		} else {
			break
		}
	}

	log.Printf("Cluster size actual %d, expected %d", actual, expected)
	assert.Equal(t, expected, actual)
	return flow
}

func (flow AcceptanceFlow) Down() AcceptanceFlow {
	flow.client.Shutdown();
	return flow.ClusterDown()
}

func (flow AcceptanceFlow) ClusterDown() AcceptanceFlow {
	err := flow.project.Down(context.Background(), options.Down{})
	if err != nil && flow.options.ImmediateFail {
		log.Fatal(err)
	}
	return flow
}

func (flow AcceptanceFlow) DefaultClient() AcceptanceFlow {
	var clientConfig = hazelcast.NewConfig()
	log.Println(flow.memberIp)
	clientConfig.NetworkConfig().SetAddresses(flow.memberIp)
	clientConfig.NetworkConfig().SetConnectionAttemptLimit(5)
	clientConfig.NetworkConfig().SetConnectionTimeout(5*time.Second)
	return flow.Client(clientConfig)
}

func (flow AcceptanceFlow) Client(config *config.Config) AcceptanceFlow {
	if flow.memberIp != nil || len(flow.memberIp) > 0 {
		config.NetworkConfig().SetAddresses(flow.memberIp)
	}

	hz_client, err := hazelcast.NewClientWithConfig(config)
	if err != nil {
		flow.ClusterDown()
		if flow.options.ImmediateFail {
			log.Fatal(err)
		}
		return flow
	}

	members := hz_client.Cluster().GetMembers()
	log.Printf("Number of members : %v", len(members))
	flow.client = hz_client
	flow.config = config
	return flow
}

func (flow AcceptanceFlow) TryMap(t *testing.T, args ...int) AcceptanceFlow {
	map_name, _ := reggen.Generate("[a-z]{42}", 42)
	mp, err := flow.client.GetMap(map_name)
	if err != nil {
		flow.Down()
		t.Fatal(err)
	}

	size, _ := mp.Size()
	assert.Equal(t, size, int32(0))

	count, valueSize := countAndSize(args...)
	if flow.options.Store {
		flow.store.mapName = map_name
	}

	samples := make([]float64, count)
	for i := 0; i < count; i++ {
		key, _ := reggen.Generate("[a-z]{42}", 42)
		value, _ := reggen.Generate("[0-9]*", valueSize)

		start := time.Now()
		mp.Put(key, value)
		actual, _ := mp.Get(key)
		end := time.Now()
		samples[i] = float64(end.Sub(start))

		assert.Equal(t, value, actual)

		if flow.options.Store {
			flow.store.entry[key] = value
		}
	}
	flow.samples = samples

	s, _ := mp.Size()
	assert.Equal(t, s, int32(count))
	if !flow.options.Store {
		mp.Clear()
		s, _ = mp.Size()
		assert.Equal(t, s, int32(0))
	}

	flow.createdMap = mp
	return flow
}

func countAndSize(args ...int) (int, int) {
	if args == nil || len(args) == 0 {
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
	if flow.createdMap == nil {
		var err error
		name, _ := reggen.Generate("[a-z]{42}", 42)
		flow.createdMap, err = flow.client.GetMap(name); if err != nil {
			log.Printf("Error is %v", err)
			return flow
		}
	}

	key, _ := reggen.Generate("[a-z]{42}", 42)
	value, _ := reggen.Generate("[a-z]{42}", 42)

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
	members := flow.client.Cluster().GetMembers()
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

	entrySet, err := flow.createdMap.EntrySetWithPredicate(predicate.Regex("this", valueRegex))
	if err != nil {
		flow.Down()
		t.Fatalf("Predicate error %v", err)
	}
	assert.Equal(t, size, len(entrySet))

	actualValues, err := flow.createdMap.ValuesWithPredicate(predicate.Regex("this", valueRegex))
	if err != nil {
		t.Fatalf("Predicate error %v", err)
	}
	assert.Equal(t, size, len(actualValues))
	assert.Subsetf(t, values, actualValues, "Fails value check")

	keySet, err := flow.createdMap.KeySetWithPredicate(predicate.Regex("this", valueRegex))
	if err != nil {
		flow.Down()
		t.Fatalf("Predicate error %v", err)
	}
	assert.Equal(t, size, len(keySet))
	assert.Subsetf(t, keys, keySet, "Fails key check")

	flow.createdMap.Clear()

	return flow
}

func (flow AcceptanceFlow) EntryProcessor(t *testing.T, expected string, processor *EntryProcessor) AcceptanceFlow {
	key, _ := reggen.Generate("^[a-z]+", 42)
	val, _ := reggen.Generate("^[0-9]+", 1024)

	_, err := flow.createdMap.Put(key, val); if err != nil {
		flow.Down()
		t.Fatal(err)
	}

	//TODO register entry processor on server also
	_, _ = flow.createdMap.ExecuteOnKey(key, processor);
	assert.Equal(t, 1, processor.count)

	flow.createdMap.Clear()
	return flow
}

type LifeCycleListener struct {
	wg        *sync.WaitGroup
	collector []string
}

func (listener *LifeCycleListener) LifecycleStateChanged(newState string) {
	listener.collector = append(listener.collector, newState)
	listener.wg.Done()
}

func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

func (flow AcceptanceFlow) ExpectConnect(t *testing.T, wg *sync.WaitGroup, listener LifeCycleListener) AcceptanceFlow {
	WaitTimeout(wg, 5000)

	msg := listener.collector[len(listener.collector) - 1]
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "STARTED")
	return flow
}

func (flow AcceptanceFlow) ExpectDisconnect(t *testing.T, wg *sync.WaitGroup, listener LifeCycleListener) AcceptanceFlow {
	WaitTimeout(wg, 5000)

	msg := listener.collector[len(listener.collector) - 1]
	assert.NotEmpty(t, msg)
	assert.Contains(t, msg, "DISCONNECTED")
	return flow
}

func (flow AcceptanceFlow) Percentile(t *testing.T, limitInMillis float64) AcceptanceFlow {
	m, _ := stats.Percentile(flow.samples, 95)
	assert.Condition(t, func() bool {
		return m <= limitInMillis * 1e6 && m > 0
	})
	return flow
}

func (flow AcceptanceFlow) VerifyStore(t *testing.T) AcceptanceFlow {
	mp, err := flow.client.GetMap(flow.store.mapName); if err != nil {
		flow.Down()
		t.Fatal(err)
	}
	for k, v := range flow.store.entry {
		actual, err := mp.Get(k)
		if err != nil {
			flow.Down()
			t.Fatal(err)
		}
		if actual != v {
			flow.Down()
			t.Fatal("key value mistmatch")
		}
	}
	return flow
}
