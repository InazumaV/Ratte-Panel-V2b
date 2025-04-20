package v2b

import "testing"

func TestPanel_GetUserList(t *testing.T) {
	r := p.GetUserList(id)
	if r.Err != nil {
		t.Fatal(r.Err)
	}
	t.Log(r.Hash)
}
