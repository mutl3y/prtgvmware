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
	"encoding/json"
	"fmt"
	ps "github.com/PRTG/go-prtg-sensor-api"
	"log"
	"sort"
	"sync"
	"time"
)

// LimitsStruct is used to set error and warning levels
type LimitsStruct struct {
	MinWarn, MaxWarn, MinErr, MaxErr, WarnMsg, ErrMsg string
}

type prtgData struct {
	mu    *sync.RWMutex
	name  string
	moid  string
	err   string
	text  string
	items []ps.SensorChannel
}

func newPrtgData(name string) *prtgData {
	p := prtgData{}
	p.items = make([]ps.SensorChannel, 0, 10)
	p.name = name
	p.mu = &sync.RWMutex{}
	return &p
}

func (p *prtgData) add(value interface{}, item ps.SensorChannel) (err error) {
	switch value.(type) {
	case float64:
		item.Value = fmt.Sprintf("%0.2f", value)
		item.Float = "1"
	default:
		item.Value = fmt.Sprintf("%v", value)
	}

	if item.Unit == "Percent" {
		item.DecimalMode = "1"
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.items = append(p.items, item)
	return
}

func (p *prtgData) print(checkTime time.Duration, txt bool) error {

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

//SensorWarn is used to return an error via PRTG message
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
