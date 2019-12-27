/*
 * Copyright © 2019.  mutl3y
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"strings"
	"sync"
)

type objData struct {
	Tags    []string
	Name    string
	RefType string
}

// TagMap holds a map of tag to object data
type TagMap struct {
	Data map[string]objData
	mu   *sync.RWMutex
}

// NewTagMap instantiates a new map to sore tag info
func NewTagMap() *TagMap {
	rtn := TagMap{}
	rtn.Data = make(map[string]objData, 0)
	rtn.mu = &sync.RWMutex{}
	return &rtn
}

func (t *TagMap) add(mo mo.Reference, tag string) {
	id := mo.Reference().Value
	t.mu.Lock()
	defer t.mu.Unlock()

	obj, ok := t.Data[id]
	if !ok {
		obj := objData{}
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
		err = c.getObjIds(tag, tm)
		if err != nil {
			return err
		}
	}

	return
}

func (c *Client) getObjIds(tag string, tm *TagMap) (err error) {
	ctx := context.Background()
	if c.r == nil {
		return fmt.Errorf("could not connect using rest client, check vcenter logs")
	}
	manager := tags.NewManager(c.r)

	workingData := make([]types.ManagedObjectReference, 0, 10)

	objs, err := manager.GetAttachedObjectsOnTags(ctx, []string{tag})

	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil
		}
		return fmt.Errorf("getObjIds issue %v", err)
	}

	if len(objs) == 0 {
		return fmt.Errorf("no results %v", err)

	}
	for _, obj := range objs[0].ObjectIDs {
		workingData = append(workingData, obj.Reference())
		rtn, err := c.getChildIds(obj.Reference())
		if err != nil {
			return fmt.Errorf("getChildIds  %v", err)
		}
		workingData = append(workingData, rtn...)
	}
	for _, obj := range workingData {
		tm.add(obj.Reference(), tag)
	}

	return nil

}

func (c *Client) getChildIds(id types.ManagedObjectReference) (rtnData []types.ManagedObjectReference, err error) {
	rtnData = make([]types.ManagedObjectReference, 0, 10)

	ctx := context.Background()
	//	ctx, _ = context.WithTimeout(ctx, 10*time.Second)
	m := view.NewManager(c.c)
	v, err := m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{id.Type}, true)
	if err != nil {
		err = fmt.Errorf("CreateContainerView %v", err)
		return nil, err

	}
	defer func() { _ = v.Destroy(ctx) }()
	switch id.Type {
	case "VirtualMachine", "Datastore", "VmwareDistributedVirtualSwitch", "DistributedVirtualPortgroup":
		return []types.ManagedObjectReference{id}, nil
	case "HostSystem":

		var wd mo.HostSystem
		err = v.Properties(ctx, id, []string{"vm", "datastore", "network"}, &wd)
		if err != nil {
			err = fmt.Errorf("vm Properties %v", err)
			return nil, err

		}
		rtnData = append(rtnData, wd.Vm...)
		rtnData = append(rtnData, wd.Network...)
		rtnData = append(rtnData, wd.Datastore...)

	case "VirtualApp":
		var wd mo.VirtualApp
		err = v.Properties(ctx, id, []string{"vm", "datastore", "network"}, &wd)
		if err != nil {
			err = fmt.Errorf("vapp Properties %v", err)
			return nil, err
		}

		rtnData = append(rtnData, wd.Vm...)
		rtnData = append(rtnData, wd.Network...)
		rtnData = append(rtnData, wd.Datastore...)

	case "ClusterComputeResource":
		var wd mo.ClusterComputeResource
		err = v.Properties(ctx, id, []string{"network", "host", "datastore"}, &wd)
		if err != nil {
			err = fmt.Errorf("cluster Properties %v", err)
			return nil, err
		}
		rtnData = append(rtnData, wd.Network...)
		rtnData = append(rtnData, wd.Host...)
		rtnData = append(rtnData, wd.Datastore...)

	case "Datacenter":
		var wd mo.Datacenter
		err = v.Properties(ctx, id, []string{"hostFolder", "datastoreFolder", "networkFolder"}, &wd)
		if err != nil {
			err = fmt.Errorf("ds Properties %v", err)
			return nil, err
		}
		x := make([]types.ManagedObjectReference, 0, 3)

		x = append(x, wd.HostFolder, wd.DatastoreFolder, wd.NetworkFolder)
		for _, v := range x {
			d, err := c.getChildIds(v)
			if err != nil {
				return nil, err
			}
			rtnData = append(rtnData, d...)
		}
	case "Folder":
		var wd mo.Folder
		err = v.Properties(ctx, id, []string{"childType", "childEntity"}, &wd)
		if err != nil {
			err = fmt.Errorf("folder properties %v", err)
			return nil, err
		}
		for _, id := range wd.ChildEntity {
			d, err := c.getChildIds(id)
			if err != nil {
				return nil, err
			}
			rtnData = append(rtnData, d...)
		}

	default:
		printJSON(false, "getChildIds, missed type, Please log an issue on Github", id.Type)
		return nil, nil
	}
	return
}
