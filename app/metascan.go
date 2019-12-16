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
	"encoding/xml"
	"fmt"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Item struct {
	Name        string `xml:"name,omitempty"`
	ID          string `xml:"id,omitempty"`
	Exefile     string `xml:"exefile"`
	Params      string `xml:"params"`
	Displayname string `xml:"displayname,attr,omitempty"`
}

type prtg struct {
	Items []Item `xml:"item"`
}

type managedObject struct {
	name, vmwareType string
}

type moidNames struct {
	moid map[string]managedObject
	mu   sync.RWMutex
}

func newMoidNames(c *Client) *moidNames {
	m, err := c.getmanagedObjectMap()
	if err != nil {
		log.Fatalf("failed to get managed object names")
	}
	mob := moidNames{
		moid: m,
		mu:   sync.RWMutex{},
	}
	return &mob
}

func (m *moidNames) GetName(moid string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.moid[moid].name

}

func (m *moidNames) Gettype(moid string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.moid[moid].vmwareType

}

func (c *Client) getmanagedObjectMap() (m map[string]managedObject, err error) {
	ctx := context.Background()
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{}, true)
	if err != nil {
		return
	}
	defer v.Destroy(ctx)

	any := []string{"ManagedEntity"}
	var objs []mo.ManagedEntity
	err = v.RetrieveWithFilter(ctx, any, []string{"name"}, &objs, nil)
	if err != nil {
		return
	}

	m = make(map[string]managedObject, 0)

	for _, v := range objs {
		m[v.Self.Value] = managedObject{
			name:       v.Name,
			vmwareType: v.Reference().Type,
		}
	}
	return
}

func (c *Client) Metascan(tags []string, tm *TagMap, Age time.Duration) (err error) {
	for _, tag := range tags {
		err := c.GetObjIds(tag, tm)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	moidNames := newMoidNames(c)

	meta, err := c.obMeta(tags, tm, moidNames, Age)
	if err != nil {
		return err
	}

	if len(meta.Items) == 0 {
		return fmt.Errorf("no data found for tags %v", tags)
	}
	output, err := xml.MarshalIndent(meta, "", "   ")
	if err != nil {
		return
	}

	fmt.Printf("%+v", string(output))

	return
}

func (c *Client) obMeta(tags []string, tm *TagMap, moidMap *moidNames, Age time.Duration) (meta prtg, err error) {

	meta = prtg{}
	meta.Items = make([]Item, 0, 10)
	for id := range tm.Data {
		na := moidMap.GetName(id)
		switch moidMap.Gettype(id) {
		case "VirtualMachine":
			meta.Items = append(meta.Items, Item{
				Name:        na,
				ID:          id,
				Exefile:     filepath.Base(os.Args[0]),
				Params:      fmt.Sprintf("summary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword --oid %v --snapAge %v", id, Age),
				Displayname: na,
			})
		case "Datastore":
			meta.Items = append(meta.Items, Item{
				Name:    "DS " + na,
				ID:      id,
				Exefile: filepath.Base(os.Args[0]),
				Params:  fmt.Sprintf("dssummary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword --oid %v", id),
			})
		case "HostSystem":
			meta.Items = append(meta.Items, Item{
				Name:    "Host " + na,
				ID:      id,
				Exefile: filepath.Base(os.Args[0]),
				Params:  fmt.Sprintf("hssummary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword --oid %v", id),
			})
		case "VmwareDistributedVirtualSwitch":
			meta.Items = append(meta.Items, Item{
				Name:    "VDS " + na,
				ID:      id,
				Exefile: filepath.Base(os.Args[0]),
				Params:  fmt.Sprintf("vdssummary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword --oid %v", id),
			})
		case "", "ClusterComputeResource", "Folder", "VirtualApp", "Datacenter", "DistributedVirtualPortgroup":
		default:
			fmt.Printf("unsupported type %v\n", moidMap.Gettype(id))
		}
	}
	sort.Slice(meta.Items, func(i, j int) bool {
		return meta.Items[i].ID < meta.Items[j].ID
	})
	return
}
