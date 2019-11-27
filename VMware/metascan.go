package VMware

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func (c *Client) Metascan(tags []string, tm *TagMap) (err error) {
	for _, tag := range tags {
		err := c.GetObjIds(tag, tm)
		if err != nil {
			log.Fatalf("%v", err)
		}
	}

	meta, err := vmMeta(tags, tm)
	if err != nil {
		return
	}

	output, err := xml.MarshalIndent(meta, "", " ")
	if err != nil {
		return
	}

	fmt.Printf("%+v", string(output))

	return
}

func vmMeta(tags []string, tm *TagMap) (meta prtg, err error) {

	meta = prtg{}
	meta.Items = make([]Item, 0, 10)
	for k := range tm.Data {

		meta.Items = append(meta.Items, Item{
			Name:    k,
			ID:      k,
			Exefile: filepath.Base(os.Args[0]),
			Params:  fmt.Sprintf("summary -U https://%%host/sdk -u %%windowsuser -p %%windowspassword -n ad"),
		})
	}
	return
}

//todo remove hardcoding from sprintf
