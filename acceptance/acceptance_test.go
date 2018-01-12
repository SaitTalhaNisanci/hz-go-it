package acceptance

import (
	"testing"
	"github.com/hazelcast/hazelcast-go-client"
	"sync"
	"github.com/lucasjones/reggen"
	"log"
	"time"
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

func TestStoreVerificationOnMultipleMemberCluster(t *testing.T) {
	flow := NewFlow()
	flow.options.Store = true
	flow.Project().Up().Scale(Scaling{Count:3}).DefaultClient().ClusterSize(t, 3).TryMap(t, 1024, 1024, 1).VerifyStore(t).Down()
}

func TestDataIntactWhenMembersDown(t *testing.T) {
	flow := NewFlow()
	flow.options.Store = true
	flow.Project().Up().DefaultClient().Scale(Scaling{Count:3}).ClusterSize(t, 3).TryMap(t, 1024, 1024).Scale(Scaling{Count:2}).ClusterSize(t, 2).VerifyStore(t).Down()
}

/**
Basic Authentication
Case 2 - Client Authentication Failure

 */
func TestClusterAuthenticationWithWrongCredentials(t *testing.T) {
	flow := NewFlow()

	name, _ := reggen.Generate("[a-z]{8}", 8)
	password, _ := reggen.Generate("[a-z]{8}", 8)

	log.Printf("name %v, password %v", name, password)
	config := hazelcast.NewHazelcastConfig()
	config.GroupConfig().SetName(name)
	config.GroupConfig().SetPassword(password)

	flow.options.ImmediateFail = false

	flow.Project().Up().Client(config)
	// unable to test anything else client panics every operation.
}

/**
Basic Config
Case 1 - Invocation Timeout/Network Config
TODO this test need to be improved for max time expectation
 */
func TestInvocationTimeout(t *testing.T) {
	flow := NewFlow()
	config := hazelcast.NewHazelcastConfig()
	config.ClientNetworkConfig().SetConnectionTimeout(1).SetRedoOperation(true).SetConnectionAttemptLimit(2).SetInvocationTimeoutInSeconds(1)

	flow = flow.Project().Up()
	time.Sleep(10 * time.Second)
	flow.Client(config).TryMap(t).Down().ExpectError(t)
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
	config.ClientNetworkConfig().SetConnectionAttemptLimit(10)

	flow.Project().Up().Client(config).TryMap(t).ClusterDown().ExpectDisconnect(t, wg, listener).Down()
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





