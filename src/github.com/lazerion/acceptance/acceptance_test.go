package acceptance

import (
	"testing"
	"github.com/hazelcast/go-client"
)

func TestSingleMemberConnection(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().DefaultClient().TryMap(t).Down()
}

func TestClusterDiscovery(t *testing.T) {
	flow := NewFlow()
	flow.Project().Up().Scale(Scaling{Count:1}).DefaultClient().TryMap(t).Down()
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



