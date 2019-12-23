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

type Devicetemplate struct {
	XMLName  xml.Name `xml:"devicetemplate"`
	Text     string   `xml:",chardata"`
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Priority string   `xml:"priority,attr"`
	Check    Check    `xml:"check,omitempty"`
	Create   []Check  `xml:"create,omitempty"`
}

func NewDeviceTemplate(Age time.Duration, Tags string) *Devicetemplate {
	d := &Devicetemplate{
		XMLName:  xml.Name{},
		ID:       "customexexml",
		Name:     "prtgvmware",
		Priority: "40",
		Check:    Check{ID: "ping", Meta: "ping"},
		Create:   make([]Check, 0, 10),
	}

	// add a ping sensor

	d.Create = append(d.Create, pingSensor, port443(), snapShotSensor(Age, Tags))
	return d
}
func (dev *Devicetemplate) add(cr Check) error {
	dev.Create = append(dev.Create, cr)
	return nil
}
func (dev *Devicetemplate) save(tplate string) (err error) {
	ou, err := xml.MarshalIndent(dev, "", "  ")
	if err != nil {
		return
	}

	outStr := xml.Header + string(ou)
	fmt.Printf("%v\n\nsaved to %v.odt\n", outStr, tplate)
	b := bytes.NewBufferString(outStr)

	return ioutil.WriteFile(tplate+".odt", b.Bytes(), os.ModePerm)
}

type Check struct {
	ID          string     `xml:"id,attr,omitempty"`
	Kind        string     `xml:"kind,attr,omitempty"`
	Meta        string     `xml:"meta,attr,omitempty"`
	Requires    string     `xml:"requires,attr,omitempty"`
	Metadata    Metadata   `xml:"metadata,omitempty"`
	Createdata  Createdata `xml:"createdata,omitempty"`
	Displayname string     `xml:"displayname,attr,omitempty"`
}
type Metadata struct {
	Exefile   string `xml:"exefile,omitempty"`
	Exeparams string `xml:"exeparams,omitempty"`
}

type Createdata struct {
	Priority           string `xml:"priority,omitempty"`
	Position           string `xml:"position,omitempty"`
	Interval           string `xml:"interval,omitempty"`
	Count              string `xml:"count,omitempty"`
	Errorintervalsdown string `xml:"errorintervalsdown,omitempty"`
	Autoacknowledge    string `xml:"autoacknowledge,omitempty"`
	Tags               string `xml:"tags,omitempty"`
	Timeout            string `xml:"timeout,omitempty"`
	Exefile            string `xml:"exefile,omitempty"`
	Exeparams          string `xml:"exeparams,omitempty"`
	Name               string `xml:"name,omitempty"`
	Mutex              string `xml:"mutexname,omitempty"`
	Decimaldigits      string `xml:"decimaldigits,omitempty"`
}

func NewCreate(id, params, tags, intervalSecs string) Check {
	c := Check{
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
			Mutex:              "prtgvmware",
		},
	}
	return c
}

var pingSensor = Check{
	ID:       "pingsensor",
	Kind:     "ping",
	Requires: "ping",
	Createdata: Createdata{
		Position: "1",
		Priority: "5",
		Interval: "5",
		Timeout:  "20",
		Count:    "5",
	},
}

func port443() Check {
	c := Check{
		ID:       "port443",
		Kind:     "ssl",
		Meta:     "port",
		Requires: "ping",
		Createdata: Createdata{
			Position: "2",
		},
	}
	return c
}

func snapShotSensor(Age time.Duration, Tags string) Check {
	name := fmt.Sprintf("snapshots older than %v hours", Age.Hours())
	c := Check{
		ID:       "snapshots",
		Kind:     "exexml",
		Meta:     "",
		Requires: "ping",
		Createdata: Createdata{Name: name, Tags: Tags, Errorintervalsdown: "5",
			Autoacknowledge: "1", Priority: "3", Exefile: filepath.Base(os.Args[0]), Mutex: "prtgvmware",
			Exeparams: fmt.Sprintf("snapshots -U https://%%Host/sdk -u %%windowsuser -p %%windowspassword --snapAge %v --tags %v --maxWarn 1 --maxErr 3", Age, Tags),
		},
	}
	return c
}

func GenTemplate(tags []string, Age time.Duration, tplate string) error {
	//fmt.Println(basetemplate)
	creds := "-U https://%Host/sdk -u %windowsuser -p %windowspassword"
	d := NewDeviceTemplate(Age, strings.Join(tags, ","))

	ch1 := fmt.Sprintf("metascan %v --snapAge %v --tags %v", creds, Age, strings.Join(tags, ","))
	ch := NewCreate("metascan", ch1, strings.Join(tags, ","), "300")
	err := d.add(ch)
	if err != nil {
		return fmt.Errorf("failed to add check %v", err)
	}

	// save to disk
	err = d.save(tplate)
	if err != nil {
		return fmt.Errorf("failed to save file %v", err)
	}

	return nil
}

func (c *Client) DynTemplate(tags []string, Age time.Duration, tplate string) error {
	d := NewDeviceTemplate(Age, strings.Join(tags, ","))

	tm := NewTagMap()
	err := c.list(tags, tm)
	for _, tag := range tags {
		err := c.GetObjIds(tag, tm)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	moidNames := newMoidNames(c)

	meta, err := c.obMeta(tm, moidNames, Age)
	if err != nil {
		return err
	}

	for _, v := range meta.Items {
		c := Check{
			ID:       v.Name,
			Kind:     "exexml",
			Meta:     "",
			Requires: "ping",
			Createdata: Createdata{Name: v.Name, Tags: strings.Join(tags, ","), Errorintervalsdown: "5",
				Autoacknowledge: "1", Priority: "3", Exefile: filepath.Base(os.Args[0]), Mutex: "prtgvmware",
				Exeparams: v.Params,
			},
		}
		err = d.add(c)
		if err != nil {
			return err
		}
	}

	// save to disk
	err = d.save(tplate)
	if err != nil {
		return fmt.Errorf("failed to save file %v", err)
	}

	return nil
}
