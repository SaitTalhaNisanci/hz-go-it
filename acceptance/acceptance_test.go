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

/**
Cluster/Lifecycle Service
Case 1 - Cluster Scale UP
 */
func TestClusterDiscoveryScaledUp(t *testing.T) {
	flow := NewFlow()
	expectedSize := 2
	flow.Project().Up().Scale(Scaling{Count:expectedSize}).DefaultClient().TryMap(t).ClusterSize(t, expectedSize).Down()
}

/**
Cluster/Lifecycle Service
Case 2 - Cluster Scale Down
 */
func TestClusterDiscoveryWhenScaledDown(t *testing.T) {
	flow := NewFlow()
	increment := 2
	expectedSize := 1
	flow.Project().Up().Scale(Scaling{Count:increment}).DefaultClient().Scale(Scaling{Count:expectedSize}).ClusterSize(t, expectedSize).Down()
}

func TestDataIntactWhenMembersDown(t *testing.T) {
	flow := NewFlow()
	flow.options.Store = true
	flow.Project().Up().Scale(Scaling{Count:3}).DefaultClient().ClusterSize(t, 3).TryMap(t, 1024, 1024).Scale(Scaling{Count:1}).ClusterSize(t, 1).VerifyStore(t).Down()
}

func TestClientWhenClusterCompletelyGoOffAndOn(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().Scale(Scaling{Count:3}).DefaultClient().ClusterSize(t, 3).Down().ClusterSize(t, 0)
	flow.Up().Scale(Scaling{Count:2}).ClusterSize(t, 2).Down()

}

/**
Basic Authentication
Case 2 - Client Authentication Failure
 */
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
	config.ClientNetworkConfig().SetConnectionTimeout(1).SetRedoOperation(true).SetConnectionAttemptLimit(1).SetInvocationTimeoutInSeconds(1)

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





