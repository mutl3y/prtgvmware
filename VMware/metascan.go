package VMware

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/vmware/govmomi/vim25/mo"
	"log"
	"os"
	"path/filepath"
	"sync"
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

func (c *Client) Metascan(tags []string, tm *TagMap, scanTypes []string) (err error) {
	for _, tag := range tags {
		err := c.GetObjIds(tag, tm)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	moidNames := newMoidNames(c)

	it := make([]Item, 0, len(scanTypes))

	for _, scan := range scanTypes {
		if scan == "vm" {
			meta, err := vmMeta(tags, tm, moidNames)
			if err != nil {
				return err
			}

			it = append(it, meta.Items...)
		}
	}

	meta := prtg{}
	meta.Items = it
	output, err := xml.MarshalIndent(meta, "", "   ")
	if err != nil {
		return
	}

	fmt.Printf("%+v", string(output))

	return
}

func vmMeta(tags []string, tm *TagMap, moidMap *moidNames) (meta prtg, err error) {

	meta = prtg{}
	meta.Items = make([]Item, 0, 10)
	for id := range tm.Data {
		na := moidMap.GetName(id)
		meta.Items = append(meta.Items, Item{
			Name:    na,
			ID:      id,
			Exefile: filepath.Base(os.Args[0]),
			Params:  fmt.Sprintf("summary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword -n %v", na),
		})
	}
	return
}

type moidNames struct {
	moid map[string]string
	mu   sync.RWMutex
}

func newMoidNames(c *Client) *moidNames {
	m, err := c.getNames()
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
	return m.moid[moid]

}

func (c *Client) getNames() (m map[string]string, err error) {
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

	m = make(map[string]string, 0)

	for _, v := range objs {
		m[v.Self.Value] = v.Name
	}
	return
}

//func (c *Client) getNameFromMoid(moid string){
//	ctx := context.Background()
//	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{}, true)
//	if err != nil {
//		return
//	}
//
//	defer v.Destroy(ctx)
//
//	any := []string{"ManagedEntity"}
//	var objs []mo.ManagedEntity
//	err = v.RetrieveWithFilter(ctx, any, []string{"name"}, &objs, nil)
//	if err != nil {
//
//		return
//	}
//
//	m := make(map[string]string,0)
//
//	for _,v := range objs{
//		m[v.Self.Value] = v.Name
//	}
//	printJson(false,m)
//
//
//
//}
////todo remove hardcoding from sprintf
