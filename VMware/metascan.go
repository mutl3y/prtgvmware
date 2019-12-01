package VMware

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Item struct {
	Name    string `xml:"name"`
	ID      string `xml:"id"`
	Exefile string `xml:"exefile"`
	Params  string `xml:"params"`
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

func (c *Client) Metascan(tags []string, tm *TagMap, scanTypes []string, Age time.Duration) (err error) {
	for _, tag := range tags {
		err := c.GetObjIds(tag, tm)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}
	moidNames := newMoidNames(c)

	meta, err := obMeta(tags, tm, moidNames, Age)
	if err != nil {
		return err
	}

	output, err := xml.MarshalIndent(meta, "", "   ")
	if err != nil {
		return
	}

	fmt.Printf("%+v", string(output))

	return
}

func obMeta(tags []string, tm *TagMap, moidMap *moidNames, Age time.Duration) (meta prtg, err error) {

	meta = prtg{}
	meta.Items = make([]Item, 0, 10)
	for id := range tm.Data {
		na := moidMap.GetName(id)
		switch moidMap.Gettype(id) {
		case "VirtualMachine":

			meta.Items = append(meta.Items, Item{
				Name:    na,
				ID:      id,
				Exefile: filepath.Base(os.Args[0]),
				Params:  fmt.Sprintf("summary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword -m %v --snapAge %v -t %v", id, Age, strings.Join(tags, ",")),
			})
		case "Datastore":
			meta.Items = append(meta.Items, Item{
				Name:    na,
				ID:      id,
				Exefile: filepath.Base(os.Args[0]),
				Params:  fmt.Sprintf("dssummary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword -m %v -t %v", id, strings.Join(tags, ",")),
			})
		default:
			fmt.Println("unsupported type", moidMap.Gettype(id))
		}
	}
	return
}
