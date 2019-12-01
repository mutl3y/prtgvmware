package VMware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	debug          bool
	summaryDefault = []string{
		"disk.read.average", "disk.write.average", "disk.usage.average",
		"cpu.readiness.average", "cpu.usage.average",
		"mem.active.average", "mem.consumed.average", "mem.usage.average",
		"net.bytesRx.average", "net.bytesTx.average", "net.usage.average",
		"datastore.datastoreNormalReadLatency.latest", "datastore.datastoreNormalWriteLatency.latest",
		"datastore.datastoreReadIops.latest", "datastore.datastoreWriteIops.latest",
	}
)

func (c *Client) VmSummary(name, moid string, lim *LimitsStruct, age time.Duration, txt bool, sensors []string) error {
	summaryDefault = append(summaryDefault, sensors...)
	start := time.Now()
	ctx := context.Background()
	if c.m == nil {
		return fmt.Errorf("no manager")
	}
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return fmt.Errorf("con view 1 %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()

	//var vms2 []mo.VirtualMachine
	//	kind := []string{"VirtualMachine"}
	vms := make([]mo.VirtualMachine, 0, 100)
	if moid != "" {
		v0 := mo.VirtualMachine{}
		err = v.Properties(ctx, types.ManagedObjectReference{
			Type:  "VirtualMachine",
			Value: moid,
		}, []string{"name", "summary", "snapshot", "guest"}, &v0)
		if err != nil {
			return err
		}
		vms = append(vms, v0)
	} else {

		err = v.RetrieveWithFilter(ctx, []string{"ManagedEntity"}, []string{"name", "summary", "snapshot", "guest"}, &vms, property.Filter{"name": name})
		if err != nil {
			return fmt.Errorf("mo match using name %v", name)
		}

	}

	if len(vms) != 1 {

		type vmFailList struct {
			name, moid string
		}
		out := make([]vmFailList, 0, 10)
		for _, v := range vms {
			out = append(out, vmFailList{v.Name, v.Self.Value})

		}

		return fmt.Errorf("expected a single vm, got %+v", out)
	}

	item := vms[0]
	//printJson(false,item)
	vm, mets, err := c.vmMetricS(item.Reference())
	if err != nil {
		return err
	}
	var co int
	if item.Snapshot != nil {
		co, err = snapshotCount(time.Now().Add(-age), item.Snapshot.RootSnapshotList)
		if err != nil {
			return err
		}
	}
	pr := NewPrtgData(item.Name)
	pr.moid = vm
	err = pr.Add(fmt.Sprintf("Snapshots Older Than %v", age), "Count", "One", co, lim, "", false)

	//	guestLimits := &LimitsStruct{
	//		MinErr: 0.5,
	//		ErrMsg: "tools not running",
	//	}
	if item.Guest.ToolsRunningStatus == "guestToolsRunning" {
		_ = pr.Add("guest tools running", "Custom", "", true, &LimitsStruct{}, "prtg.standardlookups.boolean.statetrueok", false)
	} else {
		_ = pr.Add("guest tools running", "Custom", "", false, &LimitsStruct{}, "prtg.standardlookups.boolean.statetrueok", false)

	}

	for k, v := range mets {
		if inStringSlice(k, summaryDefault) {
			st, err := singleStat(v.Value)
			if err != nil {
				return err
			}

			if st != nil {
				err = pr.Add(k, v.Unit, v.volumeSize, st, &LimitsStruct{}, "", false)
				if err != nil {
					return err
				}
			}
		}
	}
	for _, v := range item.Guest.Disk {
		d := v.DiskPath
		ca := v.Capacity
		free := v.FreeSpace
		one := ca / 100
		perc := free / one
		_ = pr.Add("free Bytes "+d, "BytesDisk", "KiloByte", free/1000, &LimitsStruct{}, "", true)
		_ = pr.Add("free Space (Percent) "+d, "Percent", "", perc, &LimitsStruct{
			MinWarn: 20,
			MinErr:  10,
			WarnMsg: "Warning Low Space",
			ErrMsg:  "Critical disk space",
		}, "", false)
	}
	err = pr.Print(time.Since(start), txt)

	return err
}

//SnapShotsOlderThan
// prints every vm unless using tags
// tag slice of tags to check
// txt to display results in json
// timeRange exclude snapshots younger than time.duration, set to 0 for all
func (c *Client) SnapShotsOlderThan(f property.Filter, tagIds []string, lim *LimitsStruct, age time.Duration, txt bool) (err error) {
	start := time.Now()
	ctx := context.Background()
	m := view.NewManager(c.c)

	v, err := m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return
	}
	defer func() { _ = v.Destroy(ctx) }()

	// retrieve snapshot info
	var vms []mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"ManagedEntity"}, []string{"snapshot", "name"}, &vms, f)
	if err != nil {
		return
	}

	// retrieve tags and object associations
	pr := NewPrtgData("snapshots")
	tm := NewTagMap()
	err = c.list(tagIds, tm)
	if (err != nil) && !strings.Contains(err.Error(), "404") {
		return
	}

	respTime := time.Since(start)

	b := time.Now().Add(-age)
	wg := sync.WaitGroup{}
	noTags := len(tagIds) == 0

	for _, v := range vms {

		wg.Add(1)
		go func(v mo.VirtualMachine) {
			defer wg.Done()
			var co int
			if v.Snapshot != nil {
				co, err = snapshotCount(b, v.Snapshot.RootSnapshotList)
				if err != nil {
					return
				}
			}

			if noTags || tm.check(v.Self.Value, tagIds) {
				stat := fmt.Sprintf("%v_%v", v.Self.Value, v.Name)
				err = pr.Add(stat, "One", "Count", co, lim, "", false)
			}
		}(v)

	}

	wg.Wait()
	err = pr.Print(respTime, txt)
	return err

}

func (c *Client) vmMetricS(mor types.ManagedObjectReference) (vm string, m map[string]Prtgitem, err error) {
	m = make(map[string]Prtgitem)

	ctx := context.Background()
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return "", nil, fmt.Errorf("con view 1 %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()

	//vmsRefs, err := v.Find(ctx, []string{"VirtualMachine"}, filter) //todo need to fix this  find does not work
	//if err != nil {
	//	return "", nil, fmt.Errorf("%v", err)
	//}
	//if len(vmsRefs) != 1 {
	//	return "", nil, fmt.Errorf("filter issue, expected 1 vm and got %v, %v", len(vmsRefs), vmsRefs)
	//}

	// Create a PerfManager
	perfManager := performance.NewManager(c.c)

	// Retrieve counters name list
	counters, err := perfManager.CounterInfoByName(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("%v", err)
	}

	var names []string
	for name := range counters {
		names = append(names, name)
	}

	// Create PerfQuerySpec
	spec := types.PerfQuerySpec{
		MaxSample:  1,
		MetricId:   []types.PerfMetricId{{Instance: "*"}},
		IntervalId: 20,
	}

	// Query metrics
	sample, err := perfManager.SampleByName(ctx, spec, names, []types.ManagedObjectReference{mor})
	if err != nil {
		return "", nil, fmt.Errorf("%v", err)
	}

	result, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return "", nil, fmt.Errorf("%v", err)
	}

	if len(result) == 0 {
		return "", nil, fmt.Errorf("no results")
	}

	//Read result
	for _, metric := range result {
		for _, v := range metric.Value {
			n := v.Name

			vm = metric.Entity.Value
			counter := counters[n]
			units := counter.UnitInfo.GetElementDescription().Label
			instance := v.Instance
			if instance == "*" {
				instance = ""
			}
			if instance != "" {
				n = fmt.Sprintf("%v.%v", n, instance)
			}

			if len(v.Value) != 0 {

				st, err := singleStat(v.ValueCSV())
				if err != nil {
					return "", nil, fmt.Errorf("singlestat failed %v", err)
				}
				if st == "-1" {
					continue
				}

				u, s := VmMetType(units, counter.GroupInfo.GetElementDescription().Key)

				m[n] = Prtgitem{
					Value:      st,
					Unit:       u,
					volumeSize: s, //todo add speedsize lookup
				}
			}
		}
	}
	return
}

func (c *Client) hostMetrics(path string) (hs string, m map[string]Prtgitem, err error) {
	m = make(map[string]Prtgitem)
	ctx := context.Background()
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return "", nil, err
	}

	defer v.Destroy(ctx)

	// Retrieve summary property for all hosts
	// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.HostSystem.html
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{}, &hss)
	if err != nil {
		return "", nil, err
	}

	b, err := json.MarshalIndent(hss[0].Summary, "", "    ")
	if err != nil {
		return "", nil, err
	}
	fmt.Printf("%+v\n", string(b))
	//
	//// Retrieve summary property for all datastores
	//// Reference: http://pubs.vmware.com/vsphere-60/topic/com.vmware.wssdk.apiref.doc/vim.Datastore.html

	//fmt.Printf("%+v\n",dss[0].)
	//
	//dsRefs, err := dv.Find(ctx, []string{"Datastore"},nil) //todo need to fix this  find does not work
	//if err != nil {
	//	return "", nil, fmt.Errorf("%v", err)
	//}
	//if len(dsRefs) != 1 {
	//	return "", nil, fmt.Errorf("filter issue, expected 1 vm and got %v, %v", len(dsRefs), dsRefs)
	//}
	//
	//var dss []mo.Datastore
	//err = dv.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	//if err != nil {
	//	return "", nil, nil
	//}
	//if len(dss) == 0 {
	//	return "", nil, fmt.Errorf("no ds returned")
	//}
	//finder := find.NewFinder(c.c,true)
	//dc,err := finder.Datacenter(ctx,"*")
	//if err != nil {
	//	return "", nil, nil
	//}
	//finder.SetDatacenter(object.NewDatacenter(c.c, dc.Reference()))
	//
	//finder.SetDatacenter(dc)
	//dso, err := finder.Datastore(ctx,"*1")
	//if err != nil {
	//	return "", nil, fmt.Errorf("dsf.datastore %v",err)
	//}
	//
	//fmt.Printf("dso %+v\n",dso)
	//srm := object.NewStorageResourceManager(c.c)
	//s, err := srm.QueryDatastorePerformanceSummary(ctx, dso)
	//fmt.Println("ds ",s)
	return hss[0].Name, nil, err
}

func unitType(s string) string {
	switch s {
	case "net":
		return "BytesBandwidth"
	case "disk", "virtualDisk", "datastore":
		return "BytesDisk"
	case "mem":
		return "BytesMemory"
	default:
		return "Custom"

	}

}

func VmMetType(u, s string) (unit, size string) {
	const (
		BytesBandwidth string = "BytesBandwidth"
		BytesDisk      string = "BytesDisk"
		Temperature    string = "Temperature"
		Percent        string = "Percent"
		TimeResponse   string = "TimeResponse"
		TimeSeconds    string = "TimeSeconds"
		Custom         string = "Custom"
		Count          string = "Count"
		CPU            string = "CPU"
		BytesFile      string = "BytesFile"
		SpeedDisk      string = "SpeedDisk"
		SpeedNet       string = "SpeedNet"
		TimeHours      string = "TimeHours"
		One            string = "One"
		Kilo           string = "Kilo"
		Mega           string = "Mega"
		Giga           string = "Giga"
		Tera           string = "Tera"
		Byte           string = "Byte"
		KiloByte       string = "KiloByte"
		MegaByte       string = "MegaByte"
		GigaByte       string = "GigaByte"
		TeraByte       string = "TeraByte"

		Bit     string = "Bit"
		KiloBit string = "KiloBit"
		MegaBit string = "MegaBit"
		GigaBit string = "GigaBit"
		TeraBit string = "TeraBit"

		Second string = "Second"
		Minute string = "Minute"
		Hour   string = "Hour"
		Day    string = "Day"
	)

	switch u {
	case "KB":
		unit = unitType(s)
		size = KiloByte

	case "MB":
		size = MegaByte
		unit = unitType(s)

	case "GB":
		size = GigaByte
		unit = unitType(s)

	case "TB":
		size = TeraByte
		unit = unitType(s)

	case "num":
		unit = Count
	case "ms":
		unit = TimeResponse

	case "%":
		unit = Percent
	case "KBps":
		size = KiloBit

		switch s {
		case "net", "mem":
			unit = "SpeedNet"
		case "disk", "virtualDisk", "datastore":
			unit = "SpeedDisk"

		default:
			unit = "Custom"
			printJson(false, "missed KBps type", s)

		}
	case "MHz":
		unit = CPU
		size = Mega

	case "℃":
		size = Temperature
	case "µs":
		size = Custom
	default:
		size = u
		//	printJson(false,u)

	}

	return
}

func snapshotCount(before time.Time, snp []types.VirtualMachineSnapshotTree) (int, error) {
	var co int
	for _, v := range snp {
		if v.CreateTime.Before(before) {
			co++
		}

		if v.ChildSnapshotList == nil {
			continue
		}
		c, err := snapshotCount(before, v.ChildSnapshotList)
		if err != nil {
			return co, err
		}
		co = c + co

	}

	return co, nil
}

func singleStat(stat interface{}) (interface{}, error) {
	var rtnStat interface{}
	switch t := stat.(type) {
	case float32, float64, int8, uint8, int16, uint16, int32, uint32, uint64, int64, uint, int, bool:
		if stat == "" {
			return nil, nil
		}

		rtnStat = stat
	case string:
		if stat == "" {
			return nil, nil
		}
		fl, err := strconv.ParseFloat(stat.(string), 64)
		if err != nil {
			fmt.Println("cant parse this ************************************************************")
			return nil, err
		}
		fl = float64(int(fl*100) / 100)
		rtnStat = fl
	case []float64:
		st := stat.([]float64)
		fl := float64(int(st[0]*100) / 100)
		rtnStat = fl
	case []int64:
		st := stat.([]int64)
		rtnStat = st[0]
	case nil:
		rtnStat = nil
	default:
		return nil, fmt.Errorf("type of %v is not supported for \n %v\n", t, stat)
	}

	return rtnStat, nil
}

func printJson(txt bool, i ...interface{}) {
	for _, v := range i {
		b, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		if txt {
			fmt.Println(i...)
		} else {

			fmt.Printf("%+v\n", string(b))
		}
	}
}

func inStringSlice(str string, strSlice []string) bool {
	for _, v := range strSlice {
		if str == v {
			return true
		}
	}
	return false
}

func tagCheck(n string, t []string) (found bool) {
	for _, check := range t {
		if n == check {
			return true
		}
	}

	return
}

func (c *Client) DsSummary(name, moid string, lim *LimitsStruct, js bool) (err error) {

	start := time.Now()
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)
	dv, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		return nil
	}

	defer dv.Destroy(ctx)

	id := types.ManagedObjectReference{
		Type:  "Datastore",
		Value: "",
	}

	if moid != "" {
		id.Value = moid
	} else {
		dss := []mo.Datastore{}
		err = dv.RetrieveWithFilter(ctx, []string{"ManagedEntity"}, []string{"name", "summary"}, &dss, property.Filter{"name": name})
		if err != nil {
			return fmt.Errorf("mo match using name %v", name)
		}
		id = dss[0].Reference()
	}

	v0 := mo.Datastore{}
	err = dv.Properties(ctx, id, []string{"name", "summary"}, &v0)
	if err != nil {
		return fmt.Errorf("ds object %v", err)
	}
	pr := NewPrtgData(v0.Name)
	pr.moid = id.Value

	whole := v0.Summary.Capacity
	free := v0.Summary.FreeSpace
	p1 := whole / 100
	freep := free / p1
	provisioned := 100 - freep
	_ = pr.Add("Available capacity", "BytesDisk", "KiloByte", whole, &LimitsStruct{}, "", false)
	_ = pr.Add("Free Bytes", "BytesDisk", "KiloByte", free, &LimitsStruct{
		MinWarn: 10,
		MinErr:  5,
		WarnMsg: "Warning Low Space",
		ErrMsg:  "Critical disk space",
	}, "", false)
	_ = pr.Add("Free space (Percent)", "Percent", "", freep, lim, "", false)
	_ = pr.Add("Provisioned", "Percent", "", provisioned, &LimitsStruct{}, "", false)
	mm := "0"
	if v0.Summary.MaintenanceMode != "normal" {
		mm = "1"
	}

	_ = pr.Add("Maintenance Mode", "Custom", "", mm, &LimitsStruct{MaxWarn: 1}, "prtg.standardlookups.boolean.statefalseok", false)
	//dsRefs, err := dv.Find(ctx, []string{"Datastore"},nil) //todo need to fix this  find does not work
	//if err != nil {
	//	return "", nil, fmt.Errorf("%v", err)
	//}
	//if len(dsRefs) != 1 {
	//	printJson(false,dsRefs)
	//
	//	return "", nil, fmt.Errorf("filter issue, expected 1 vm and got %v, %v", len(dsRefs), dsRefs)
	//}
	//
	//var dss []mo.Datastore
	//err = dv.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	//if err != nil {
	//	return "", nil, nil
	//}
	//if len(dss) == 0 {
	//	return "", nil, fmt.Errorf("no ds returned")
	//}
	//
	//finder := find.NewFinder(c.c, true)
	//dc, err := finder.Datacenter(ctx, "*")
	//if err != nil {
	//	return "", nil, nil
	//}
	//finder.SetDatacenter(object.NewDatacenter(c.c, dc.Reference()))
	//
	//finder.SetDatacenter(dc)
	//dso, err := finder.Datastore(ctx, "*1")
	//if err != nil {
	//	return "", nil, fmt.Errorf("dsf.datastore %v", err)
	//}
	//
	////fmt.Printf("dso %+v\n", dso)
	////srm := object.NewStorageResourceManager(c.c)
	////s, err := srm.QueryDatastorePerformanceSummary(ctx, dso)
	////printJson(false, s)
	//return dso.Name(), nil, err
	err = pr.Print(time.Since(start), js)
	return nil
}
