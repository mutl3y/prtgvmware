package VMware

import (
	"context"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	"sync"
)

type obJdata struct {
	Tags    []string
	Name    string
	RefType string
}

type TagMap struct {
	Data map[string]obJdata
	mu   *sync.RWMutex
}

func NewTagMap() *TagMap {
	rtn := TagMap{}
	rtn.Data = make(map[string]obJdata, 0)
	rtn.mu = &sync.RWMutex{}
	return &rtn
}

func (t *TagMap) add(mo mo.Reference, tag string) {
	id := mo.Reference().Value
	t.mu.Lock()
	defer t.mu.Unlock()

	obj, ok := t.Data[id]
	if !ok {
		obj := obJdata{}
		obj.Tags = make([]string, 0, 10)
	}
	obj.RefType = mo.Reference().Type
	obj.Tags = append(obj.Tags, tag)
	t.Data[id] = obj
}

func (t *TagMap) check(vm string, tag []string) bool {
	t.mu.RLock()
	td := t.Data[vm].Tags
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

func (c *Client) list(tagIds []string, tm *TagMap) (err error) {

	for _, tag := range tagIds {
		err = c.GetObjIds(tag, tm)
	}

	return
}

func (c *Client) GetObjIds(tag string, tm *TagMap) (err error) {
	ctx := context.Background()
	manager := tags.NewManager(c.r)

	objs, err := manager.GetAttachedObjectsOnTags(ctx, []string{tag})
	if err == nil {
		if len(objs) > 0 {
			for _, obj := range objs[0].ObjectIDs {
				tm.add(obj.Reference(), tag)
			}
		}

	}
	return
}
