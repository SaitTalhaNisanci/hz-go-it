package acceptance

import (
	"testing"
)

func TestSingleMemberConnection(t *testing.T) {
	flow := NewFlow()
	flow.Project("./deployment.yaml").Up().Client().TryMap(t).Down()
}

