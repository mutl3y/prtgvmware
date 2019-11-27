package VMware

import (
	"context"
	"github.com/vmware/govmomi/vapi/tags"
	"sync"
)

type TagMap struct {
	Data map[string][]string
	mu   *sync.RWMutex
}

func NewTagMap() *TagMap {
	rtn := TagMap{}
	rtn.Data = make(map[string][]string, 0)
	rtn.mu = &sync.RWMutex{}
	return &rtn
}

func (t *TagMap) add(vm, tag string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Data[vm] = append(t.Data[vm], tag)
}

func (t *TagMap) check(vm string, tag []string) bool {
	t.mu.RLock()
	td := t.Data[vm]
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

func (c *Client) tagList(tagIds []string, tm *TagMap) (err error) {

	for _, tag := range tagIds {
		err = c.GetObjIds(tag, tm)
	}

	return
}

func (c *Client) GetObjIds(tag string, tm *TagMap) (err error) {
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
