package acceptance

import (
	"testing"
)

func TestSingleMemberConnection(t *testing.T) {
	flow := NewFlow()
	flow.Project("./deployment.yaml").Up().Client().TryMap(t).Down()
}

func TestClusterDiscovery(t *testing.T)  {
	flow := NewFlow()
	flow.Project("./deployment.yaml").Up().Scale().Client().Down()
}

