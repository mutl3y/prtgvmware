package VMware

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var filename = filepath.Base(os.Args[0])

type Create struct {
	ID       string `xml:"id,attr"`
	Kind     string `xml:"kind,attr"`
	Meta     string `xml:"meta,attr"`
	Requires string `xml:"requires,attr"`
	Metadata struct {
		Exefile   string `xml:"exefile"`
		Exeparams string `xml:"exeparams"`
	} `xml:"metadata"`
	Createdata struct {
		Priority string `xml:"priority"`
		Interval string `xml:"interval"`
		Tags     string `xml:"tags"`
	} `xml:"createdata"`
}

func NewCreate(id, params, tags, intervalSecs string) Create {
	c := Create{
		ID:       id,
		Kind:     "exexml",
		Meta:     "customexexmlscan",
		Requires: "ping",
		Metadata: struct {
			Exefile   string `xml:"exefile"`
			Exeparams string `xml:"exeparams"`
		}{
			Exefile:   filename,
			Exeparams: params,
		},
		Createdata: struct {
			Priority string `xml:"priority"`
			Interval string `xml:"interval"`
			Tags     string `xml:"tags"`
		}{
			Priority: "4",
			Interval: intervalSecs,
			Tags:     tags,
		},
	}
	return c
}

type Devicetemplate struct {
	XMLName  xml.Name `xml:"devicetemplate"`
	Text     string   `xml:",chardata"`
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Priority string   `xml:"priority,attr"`
	Check    struct {
		Text string `xml:",chardata"`
		ID   string `xml:"id,attr"`
		Meta string `xml:"meta,attr"`
	} `xml:"check"`
	Create []Create `xml:"create"`
}

func NewDeviceTemplate() *Devicetemplate {
	d := &Devicetemplate{
		XMLName:  xml.Name{},
		Text:     "",
		ID:       "customexexml",
		Name:     "prtgvmware",
		Priority: "4",
		Check: struct {
			Text string `xml:",chardata"`
			ID   string `xml:"id,attr"`
			Meta string `xml:"meta,attr"`
		}{"", "ping", "ping"},
		Create: make([]Create, 0, 10),
	}
	return d
}

func (dev *Devicetemplate) add(cr Create) error {
	dev.Create = append(dev.Create, cr)
	return nil
}
func (dev *Devicetemplate) save() (err error) {
	ou, err := xml.MarshalIndent(dev, "", "    ")
	if err != nil {
		return
	}

	outStr := xml.Header + string(ou)
	fmt.Printf("saving the following to prtgvmware.odt\n\n%v", outStr)
	b := bytes.NewBufferString(outStr)

	return ioutil.WriteFile("prtgvmware.odt", b.Bytes(), os.ModePerm)
}

func GenTemplate(tags []string) error {
	//fmt.Println(basetemplate)

	d := NewDeviceTemplate()

	ch := NewCreate("VM Summary", "metascan -U https://%host/sdk -u %windowsuser -p %windowspassword", strings.Join(tags, ","), "60")

	err := d.add(ch)
	if err != nil {
		return fmt.Errorf("failed to add check %v", err)
	}

	err = d.save()
	if err != nil {
		return fmt.Errorf("failed to save file %v", err)
	}

	return nil
}
