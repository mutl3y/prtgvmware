package VMware

import (
	"encoding/json"
	"fmt"
	"github.com/PRTG/go-prtg-sensor-api"
	"sync"
	"time"
)

type Prtgitem struct {
	Value            interface{}
	Unit             string
	WarnMsg, ErrMsg  string
	MinWarn, MaxWarn float64
	MinErr, MaxErr   float64
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

func (p *PrtgData) Add(name, unit string, value interface{}, lim *LimitsStruct) error {
	st, err := singleStat(value)
	if err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	pi := Prtgitem{
		Value: st,
		Unit:  unit,
	}
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

	p.items[name] = pi
	return nil
}

func (p *PrtgData) Get(name string) interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.items[name]
}

func (p *PrtgData) Print(start time.Time, txt bool) error {
	checkTime := time.Since(start)

	s := prtg.New()
	if p.err != "" {
		s.SetError(true)
		s.SetSensorText(p.err)
		js, err := s.MarshalToString()
		if err != nil {
			return fmt.Errorf("marshal to string %v", err)
		}
		fmt.Println(js)
		return fmt.Errorf("error state %v", p.err)
	}
	for k, v := range p.items {

		c := s.AddChannel(k).SetValue(v.Value)
		c.Unit = v.Unit
		if v.ErrMsg != "" {
			c.LimitErrorMsg = fmt.Sprintf("%v", v.ErrMsg)
		}
		if v.WarnMsg != "" {
			c.LimitWarningMsg = fmt.Sprintf("%v", v.WarnMsg)
		}
		if v.MinErr != 0 {
			c.LimitMinError = fmt.Sprintf("%v", v.MinErr)
		}
		if v.MaxErr != 0 {
			fmt.Println("triggered")
			fmt.Printf("%T %v\n", v.MaxErr, v.MaxErr)
			c.LimitMaxError = fmt.Sprintf("%v", v.MaxErr)
		}
		if v.MinWarn != 0 {
			c.LimitMinWarning = fmt.Sprintf("%v", v.MinWarn)
		}
		if v.MaxWarn != 0 {
			c.LimitMaxWarning = fmt.Sprintf("%v", v.MaxWarn)
		}

	}

	// Response time channel
	s.AddChannel("Execution time").SetValue(checkTime.Seconds() * 1000).SetUnit(prtg.TimeResponse)

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
	fmt.Println(p.name)
	fmt.Printf("%+v\n", string(b))

	return nil
}
