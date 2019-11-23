package VMware

import (
	"context"
	"github.com/vmware/govmomi/vapi/tags"
	"sync"
)

type tagMap struct {
	data map[string][]string
	mu   *sync.RWMutex
}

func newTagMap() *tagMap {
	rtn := tagMap{}
	rtn.data = make(map[string][]string, 0)
	rtn.mu = &sync.RWMutex{}
	return &rtn
}

func (t *tagMap) add(vm, tag string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.data[vm] = append(t.data[vm], tag)
}

func (t *tagMap) check(vm string, tag []string) bool {
	t.mu.RLock()
	td := t.data[vm]
	t.mu.RUnlock()
	for _, v := range td {
		for _, v2 := range tag {
			if v2 == v {
				return true
			}
		}

	}
	return false
}

func (c *Client) tagList(tagIds []string, tm *tagMap) (err error) {

	for _, tag := range tagIds {
		err = c.getObjIds(tag, tm)
	}

	return
}

func (c *Client) getObjIds(tag string, tm *tagMap) (err error) {
	ctx := context.Background()
	manager := tags.NewManager(c.r)
	vms, err := manager.GetAttachedObjectsOnTags(ctx, []string{tag})
	if err == nil {
		for _, vm := range vms[0].ObjectIDs {
			tm.add(vm.Reference().Value, tag)
		}
	}
	return
}
