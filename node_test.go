package v2b

import "testing"

func TestPanel_GetNodeInfo(t *testing.T) {
	r := p.GetNodeInfo(id)
	if r.Err != nil {
		t.Fatal(r.Err)
	}
	t.Log(r.NodeInfo)
}
