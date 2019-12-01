package VMware

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var filename = filepath.Base(os.Args[0])

type Create struct {
	ID         string     `xml:"id,attr"`
	Kind       string     `xml:"kind,attr"`
	Meta       string     `xml:"meta,attr,omitempty"`
	Requires   string     `xml:"requires,attr"`
	Metadata   Metadata   `xml:"metadata,omitempty"`
	Createdata Createdata `xml:"createdata"`
}
type Metadata struct {
	Exefile   string `xml:"exefile,omitempty"`
	Exeparams string `xml:"exeparams,omitempty"`
}

type Createdata struct {
	Priority           string `xml:"priority,omitempty"`
	Interval           string `xml:"interval,omitempty"`
	Count              string `xml:"count,omitempty"`
	Errorintervalsdown string `xml:"errorintervalsdown,omitempty"`
	Autoacknowledge    string `xml:"autoacknowledge,omitempty"`
	Tags               string `xml:"tags,omitempty"`
	Timeout            string `xml:"timeout,omitempty"`
	Exefile            string `xml:"exefile,omitempty"`
	Exeparams          string `xml:"exeparams,omitempty"`
	Name               string `xml:"name,omitempty"`
}

type Check struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
	Meta string `xml:"meta,attr"`
}

func NewCreate(id, params, tags, intervalSecs string) Create {
	c := Create{
		ID:       id,
		Kind:     "exexml",
		Meta:     "customexexmlscan",
		Requires: "ping",
		Metadata: Metadata{
			Exefile:   filename,
			Exeparams: params,
		},
		Createdata: Createdata{
			Priority:           "3",
			Interval:           intervalSecs,
			Tags:               tags,
			Errorintervalsdown: "5",
			Autoacknowledge:    "1",
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

func NewDeviceTemplate(Age time.Duration, Tags string) *Devicetemplate {
	d := &Devicetemplate{
		XMLName:  xml.Name{},
		Text:     "",
		ID:       "customexexml",
		Name:     "prtgvmware",
		Priority: "3",
		Check:    Check{"", "ping", "ping"},
		Create:   make([]Create, 0, 10),
	}

	// add a ping sensor

	d.Create = append(d.Create, pingSensor, snapShotSensor(Age, Tags))
	return d
}

var pingSensor = Create{
	ID:       "pingsensor",
	Kind:     "ping",
	Requires: "ping",
	Createdata: Createdata{
		Priority: "5",
		Interval: "5",
		Timeout:  "20",
		Count:    "5",
	},
}

func snapShotSensor(Age time.Duration, Tags string) Create {
	name := fmt.Sprintf("snapshots older than %v hours", Age.Hours())
	c := Create{
		ID:       "snapshots",
		Kind:     "exexml",
		Meta:     "",
		Requires: "ping",
		Createdata: Createdata{Name: name, Tags: Tags, Errorintervalsdown: "5", Autoacknowledge: "1", Priority: "2", Exefile: filepath.Base(os.Args[0]),
			Exeparams: fmt.Sprintf("snapshots -U https://%%host/sdk -u %%windowsuser -p %%windowspassword --snapAge %v -t %v --MaxWarn 1 --MaxErr 3", Age, Tags),
		},
	}
	return c
}

func (dev *Devicetemplate) add(cr Create) error {
	dev.Create = append(dev.Create, cr)
	return nil
}
func (dev *Devicetemplate) save() (err error) {
	ou, err := xml.MarshalIndent(dev, "", "  ")
	if err != nil {
		return
	}

	outStr := xml.Header + string(ou)
	fmt.Printf("saving the following to prtgvmware.odt\n\n%v", outStr)
	b := bytes.NewBufferString(outStr)

	return ioutil.WriteFile("prtgvmware.odt", b.Bytes(), os.ModePerm)
}

func GenTemplate(tags []string, Age time.Duration) error {
	//fmt.Println(basetemplate)
	creds := "-U https://%host/sdk -u %windowsuser -p %windowspassword"
	d := NewDeviceTemplate(Age, strings.Join(tags, ","))

	ch1 := fmt.Sprintf("metascan %v --snapAge %v -t %v", creds, Age, strings.Join(tags, ","))
	ch := NewCreate("metascan", ch1, strings.Join(tags, ","), "60")
	err := d.add(ch)
	if err != nil {
		return fmt.Errorf("failed to add check %v", err)
	}

	// save to disk
	err = d.save()
	if err != nil {
		return fmt.Errorf("failed to save file %v", err)
	}

	return nil
}
