package VMware

import (
	"encoding/json"
	"fmt"
	prtgSensor "github.com/PRTG/go-prtg-sensor-api"
	"log"
	"sort"
	"sync"
	"time"
)

type Prtgitem struct {
	Value            interface{}
	Unit             string
	volumeSize       string
	Lookup           string
	WarnMsg, ErrMsg  string
	MinWarn, MaxWarn float64
	MinErr, MaxErr   float64
	Hide             bool
}

type PrtgData struct {
	mu    *sync.RWMutex
	name  string
	moid  string
	err   string
	items map[string]Prtgitem
}

//func (p *PrtgData) Len() int {
//	return len(p.items)
//}
//
//func (p *PrtgData) Less(i, j int) bool {
//	return p.items[i].Name < p.items[j].Name
//}
//
//func (p *PrtgData) Swap(i, j int) {
//	p.items[i], p.items[j] = p.items[j], p.items[i]
//}

func NewPrtgData(name string) *PrtgData {
	p := PrtgData{}
	p.items = make(map[string]Prtgitem)
	p.name = name
	p.mu = &sync.RWMutex{}
	return &p
}

type LimitsStruct struct {
	MinWarn, MaxWarn float64
	MinErr, MaxErr   float64
	WarnMsg, ErrMsg  string
}

func (p *PrtgData) Add(name, unit, volumeSize string, value interface{}, lim *LimitsStruct, lookup string, hide bool) error {

	st, err := singleStat(value)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	pi := Prtgitem{
		Value:      st,
		Unit:       unit,
		volumeSize: volumeSize,
	}
	pi.Hide = hide
	if lim.ErrMsg != "" {
		pi.ErrMsg = lim.ErrMsg
	}
	if lim.MinWarn != 0 {
		pi.MinWarn = lim.MinWarn
	}
	if lim.MaxWarn != 0 {
		pi.MaxWarn = lim.MaxWarn
	}

	pi.WarnMsg = lim.WarnMsg

	pi.ErrMsg = lim.ErrMsg

	if lim.MinErr != 0 {
		pi.MinErr = lim.MinErr
	}
	if lim.MaxErr != 0 {
		pi.MaxErr = lim.MaxErr
	}

	pi.Lookup = lookup

	p.items[name] = pi
	return nil
}

func (p *PrtgData) Get(name string) interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.items[name]
}

func (p *PrtgData) Print(checkTime time.Duration, txt bool) error {

	s := prtgSensor.New()
	if p.err != "" {
		SensorWarn(fmt.Errorf("%v", p.err), true)

		return fmt.Errorf("error state %v", p.err)
	}
	keys := make([]string, 0, len(p.items))
	for k := range p.items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		item := p.items[k]
		c := s.AddChannel(k).SetValue(item.Value)
		c.Unit = item.Unit
		c.VolumeSize = item.volumeSize
		if item.ErrMsg != "" {
			c.LimitErrorMsg = fmt.Sprintf("%v", item.ErrMsg)
			c.LimitMode = "1"
		}
		if item.WarnMsg != "" {
			c.LimitWarningMsg = fmt.Sprintf("%v", item.WarnMsg)
			c.LimitMode = "1"

		}
		if item.MinErr != 0 {
			c.LimitMinError = fmt.Sprintf("%v", item.MinErr)
			c.LimitMode = "1"

		}
		if item.MaxErr != 0 {
			c.LimitMaxError = fmt.Sprintf("%v", item.MaxErr)
			c.LimitMode = "1"
		}
		if item.MinWarn != 0 {
			c.LimitMinWarning = fmt.Sprintf("%v", item.MinWarn)
			c.LimitMode = "1"

		}
		if item.MaxWarn != 0 {
			c.LimitMaxWarning = fmt.Sprintf("%v", item.MaxWarn)
			c.LimitMode = "1"
		}
		//if inStringSlice(c.Unit, []string{"Byte"}) {
		//	c.VolumeSize = "1"
		//}
		//if inStringSlice(c.Unit, []string{"Bit"}) {
		//	c.SpeedSize = "1"
		//}
		c.ShowChart = "1"
		c.ShowTable = "1"
		if item.Hide {
			c.ShowChart = "0"
			c.ShowTable = "0"
		}
		if item.Lookup != "" {
			c.ValueLookup = item.Lookup

		}
	}

	// Response time channel
	s.AddChannel("Execution time").SetValue(checkTime.Seconds() * 1000).SetUnit(prtgSensor.TimeResponse)

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
	//	fmt.Println(p.name)
	fmt.Printf("%+v\n", string(b))

	return nil
}

func SensorWarn(inErr error, er bool) {
	s := prtgSensor.New()
	if er {
		s.SetError(true)
	} else {
		s.SetError(false)
		c := s.AddChannel("Execution time").SetValue(999).SetUnit(prtgSensor.TimeResponse)
		c.Warning = "1"

	}
	s.SetSensorText(inErr.Error())
	js, err := s.MarshalToString()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(js)

}
