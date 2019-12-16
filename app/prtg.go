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
	"crypto/tls"
	"encoding/json"
	"fmt"
	ps "github.com/PRTG/go-prtg-sensor-api"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

// ******* from here ************ //
type LimitsStruct struct {
	MinWarn, MaxWarn, MinErr, MaxErr, WarnMsg, ErrMsg string
}

type PrtgData struct {
	mu    *sync.RWMutex
	name  string
	moid  string
	err   string
	text  string
	items []ps.SensorChannel
}

func NewPrtgData(name string) *PrtgData {
	p := PrtgData{}
	p.items = make([]ps.SensorChannel, 0, 10)
	p.name = name
	p.mu = &sync.RWMutex{}
	return &p
}

func (p *PrtgData) Add(value interface{}, item ps.SensorChannel) (err error) {
	switch value.(type) {
	case float64:
		item.Value = fmt.Sprintf("%0.2f", value)
		item.Float = "1"
	default:
		item.Value = fmt.Sprintf("%v", value)
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.items = append(p.items, item)
	return
}

func (p *PrtgData) Print(checkTime time.Duration, txt bool) error {

	s := ps.New()
	if p.err != "" {
		SensorWarn(fmt.Errorf("%v", p.err), true)

		return fmt.Errorf("error state %v", p.err)
	}
	//keys := make([]string, 0, len(p.items))
	//for k := range p.items {
	//	keys = append(keys, k)
	//}
	//sort.Strings(keys)
	//var text string

	sort.Slice(p.items, func(i, j int) bool {
		return p.items[i].Channel <= p.items[j].Channel
	})

	for _, item := range p.items {
		sc := s.AddChannel(item.Channel)
		*sc = item
	}

	//	c.Unit = item.Unit
	//	c.VolumeSize = item.volumeSize
	//	if item.ErrMsg != "" {
	//		c.LimitErrorMsg = fmt.Sprintf("%v", item.ErrMsg)
	//		c.LimitMode = "1"
	//	}
	//	if item.WarnMsg != "" {
	//		c.LimitWarningMsg = fmt.Sprintf("%v", item.WarnMsg)
	//		c.LimitMode = "1"
	//
	//	}
	//	if item.MinErr != 0 {
	//		c.LimitMinError = fmt.Sprintf("%v", item.MinErr)
	//		c.LimitMode = "1"
	//
	//	}
	//	if item.MaxErr != 0 {
	//		c.LimitMaxError = fmt.Sprintf("%v", item.MaxErr)
	//		c.LimitMode = "1"
	//	}
	//	if item.MinWarn != 0 {
	//		c.LimitMinWarning = fmt.Sprintf("%v", item.MinWarn)
	//		c.LimitMode = "1"
	//
	//	}
	//	if item.MaxWarn != 0 {
	//		c.LimitMaxWarning = fmt.Sprintf("%v", item.MaxWarn)
	//		c.LimitMode = "1"
	//	}
	//	//if inStringSlice(c.Unit, []string{"Byte"}) {
	//	//	c.VolumeSize = "1"
	//	//}
	//	//if inStringSlice(c.Unit, []string{"Bit"}) {
	//	//	c.SpeedSize = "1"
	//	//}
	//	c.ShowChart = "1"
	//	c.ShowTable = "1"
	//	if item.Hide {
	//		c.ShowChart = "0"
	//		c.ShowTable = "0"
	//	}
	//	if item.Lookup != "" {
	//		c.ValueLookup = item.Lookup
	//	}
	//
	//}
	s.SetSensorText(p.text)
	s.SetError(false)

	// Response time channel
	s.AddChannel("Execution time").SetValue(checkTime.Seconds() * 1000).SetUnit(ps.TimeResponse)

	if !txt {
		js, err := s.MarshalToString()
		if err != nil {
			return fmt.Errorf("prtgdata.print marshall to string %v", err)
		}

		fmt.Println(js)
		return nil
	}

	b, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}
	fmt.Println(p.name, p.moid)
	fmt.Printf("%+v\n", string(b))

	return nil
}

func SensorWarn(inErr error, er bool) {
	s := ps.New()
	if er {
		s.SetError(true)
	} else {
		s.SetError(false)
		c := s.AddChannel("Execution time").SetValue(999).SetUnit(ps.TimeResponse)
		c.Warning = "1"

	}
	s.SetSensorText(inErr.Error())
	js, err := s.MarshalToString()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(js)

}

func PostComment(sensorId, comment string) error {
	//tr := http.DefaultTransport.(*http.Transport)
	//tr.TLSClientConfig = &tls.Config{}
	//tr.TLSClientConfig.InsecureSkipVerify = true

	//client := http.Client{
	//	Transport:     tr,
	//	CheckRedirect: nil,
	//	Jar:           nil,
	//	Timeout:       20,
	//}
	//req := newhttp

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	_, err := http.Get("https://golang.org/")
	if err != nil {
		fmt.Println(err)
	}
	userCreds := `&username=vmsummary&passhash=2812313609`

	resp, err := http.Post(`http://192.168.0.4/api/acknowledgealarm.htm?id=2437&ackmsg=Ticket%20Created%20-%20123456`+userCreds, "application/json", nil)
	if err != nil {
		printJson(false, "https://192.168.0.4/api/setobjectproperty.htm?id=2437&name=mutex&value=192.168.59.2"+userCreds)
		return err
	}
	fmt.Println(resp)
	return nil
}
