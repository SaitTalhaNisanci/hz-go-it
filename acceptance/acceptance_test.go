package acceptance

import (
	"testing"
	"github.com/hazelcast/go-client"
	"sync"
	"github.com/lucasjones/reggen"
)

func TestSingleMemberConnection(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().DefaultClient().TryMap(t).Down()
}

func TestClusterDiscovery(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().Scale(Scaling{Count:1}).DefaultClient().TryMap(t).Down()
}

func TestClusterAuthenticationWithWrongCredentials(t *testing.T) {
	flow := NewFlow()

	name, _ := reggen.Generate("[a-z]42", 42)
	password, _ := reggen.Generate("[a-z]42", 42)

	config := hazelcast.NewHazelcastConfig()
	config.GroupConfig().SetName(name)
	config.GroupConfig().SetPassword(password)

	flow.options.ImmediateFail = false

	flow.Project().Up().Client(config).ExpectError(t).Down()
}

/**
Basic Config
Case 1 - Invocation Timeout/Network Config
 */
func TestInvocationTimeout(t *testing.T) {
	flow := NewFlow()
	config := hazelcast.NewHazelcastConfig()
	config.ClientNetworkConfig().SetRedoOperation(true).SetConnectionAttemptLimit(1).SetInvocationTimeoutInSeconds(1)

	flow.Project().Up().Client(config).TryMap(t).Down().ExpectError(t)
}

/**
Basic Config
Case 2 - Smart Routing When Set to False
 */
func TestSmartRoutingDisabled(t *testing.T) {
	flow := NewFlow()
	config := hazelcast.NewHazelcastConfig()
	config.ClientNetworkConfig().SetSmartRouting(false)

	flow.Project().Up().Client(config).TryMap(t).ExpectConnection(t, 1).Down()
}

/**
Basic Map
Case 1 - Basic Map Get/Put/Delete
 */
func TestBasicMapOperations(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().DefaultClient().TryMap(t, 1024, 1024).Down()
}

/**
Predicate
Case 1 - Basic Map Get/Put/Delete
 */
func TestPredicate(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().DefaultClient().TryMap(t).Predicate(t).Down()
}

/**
Entry processor
Case 1 - Basic Map Get/Put/Delete
 */
func TestEntryProcessor(t *testing.T) {
	flow := NewFlow()
	config := hazelcast.NewHazelcastConfig()
	expected := "test"
	processor := CreateEntryProcessor(expected)

	config.SerializationConfig().AddDataSerializableFactory(processor.identifiedFactory.factoryId, processor.identifiedFactory)
	flow.Project().Up().Client(config).TryMap(t).EntryProcessor(t, expected, processor).Down()
}

/**
Cluster/Lifecycle Service
Case 4 - Lifecycle When Scale UP
 */
func TestWhenClusterScaleUp(t *testing.T) {
	flow := NewFlow()

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(10)
	config := hazelcast.NewHazelcastConfig()
	listener := LifeCycleListener{wg: wg, collector: make([]string, 0)}
	config.AddLifecycleListener(&listener)
	config.ClientNetworkConfig().SetConnectionAttemptLimit(5)

	flow.Project().Up().Client(config).TryMap(t).ExpectConnect(t, wg, listener).Down()
}

/**
Cluster/Lifecycle Service
Case 3 - Lifecycle When Scale DOWN
 */
func TestWhenClusterScaleDown(t *testing.T) {
	flow := NewFlow()

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	wg.Add(10)
	config := hazelcast.NewHazelcastConfig()
	listener := LifeCycleListener{wg: wg, collector: make([]string, 0)}
	config.AddLifecycleListener(&listener)
	config.ClientNetworkConfig().SetConnectionAttemptLimit(1)

	flow.Project().Up().Client(config).TryMap(t).Down().ExpectDisconnect(t, wg, listener)
}

/**
Performance Tests
Case 1 - Basic Map Operations
Note: current 95 percentile is slightly above 2 ms
 */
func TestBasicMapOperationsPerformance(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().DefaultClient().TryMap(t, 1024, 1024).Percentile(t, 3).Down()
}





