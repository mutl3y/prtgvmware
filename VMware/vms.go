package VMware

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"strings"
	"time"
)

var (
	debug          bool
	summaryDefault = []string{
		"disk.read.average", "disk.write.average", "disk.usage.average",
		"cpu.ready.summation", "cpu.usage.average", "cpu.ready.summation",
		"mem.active.average", "mem.consumed.average", "mem.usage.average",
		"net.bytesRx.average", "net.bytesTx.average", "net.usage.average",
		"datastore.datastoreNormalReadLatency.latest", "datastore.datastoreNormalWriteLatency.latest",
		"datastore.datastoreReadIops.latest", "datastore.datastoreWriteIops.latest",
	}
)

//VmSummary
// Takes a VMWare property filter such as property.filter{"name":"*vm1"}
// txt to display results in json
// age exclude snapshots younger than time.duration
func (c *Client) VmSummary(f property.Filter, lim *LimitsStruct, age time.Duration, txt bool) error {
	start := time.Now()
	ctx := context.Background()

	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return fmt.Errorf("con view 1 %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()

	var vms2 []mo.VirtualMachine
	//err = v.Retrieve(ctx, []string{"name"}, []string{}, &vms2)
	kind := []string{"VirtualMachine"}

	err = v.RetrieveWithFilter(ctx, kind, []string{"name", "summary", "snapshot"}, &vms2, f)

	printJson()
	if err != nil {
		if err.Error() == "object references is empty" {
			var s string
			for k, v := range f {
				if s != "" {
					s = s + ","
				}
				s = s + k + "=" + v.(string)
			}
			return fmt.Errorf("mo match using property filter %v", s)
		}

		return fmt.Errorf("vmsummary retrieve %v %v", f, err)
	}

	if len(vms2) != 1 {
		type vmFailList struct {
			name, moid string
		}
		out := make([]vmFailList, 0, 10)
		for _, v := range vms2 {
			out = append(out, vmFailList{v.Name, v.Self.Value})

		}

		return fmt.Errorf("expected a single vm, got %+v", out)
	}

	item := vms2[0]
	//	printJson(item)
	vm, mets, err := c.vmMetricS(f)
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
	err = pr.Add(fmt.Sprintf("Snapshots Older Than %v", age), "One", co, lim)

	for _, v := range summaryDefault {
		st, err := singleStat(mets[v].Value)
		if err != nil {
			return err
		}
		if st != nil {
			err = pr.Add(v, mets[v].Unit, st, &LimitsStruct{})

			if err != nil {
				return err
			}
		}
	}

	if debug {
		for k, v := range mets {
			if strings.Contains(k, "datastore") {
				fmt.Println(k, v.Value, v.Unit)
			}

		}
	}

	//for _, v := range vms2 {
	//	//fmt.Printf("name %v\n",v.Summary.Vm)
	//	//fmt.Printf("managed identity %v\n",v.ManagedEntity)
	//	////fmt.Printf("guest %v\n",v.Guest)
	//	//fmt.Printf("cap %v\n",v.Capability)
	//	//fmt.Printf("conf %v\n",v.Config)
	//	//fmt.Printf("task %v\n",v.Entity())
	//	//fmt.Printf("storage %v\n",v.Storage)
	//	//fmt.Printf("summary %v\n",v.Summary)
	//	//fmt.Printf("snapshot %v\n",v.Snapshot)
	//	//fmt.Printf("ref %+v\n",v.Reference())
	//	//fmt.Printf("everything\n %+v\n",v)
	//	b := time.Now()
	//	co, err := snapshotCount(b, v.Snapshot.RootSnapshotList)
	//	if err != nil {
	//		return fmt.Errorf("snapcount 1 %v", err)
	//	}
	//
	//	fmt.Printf("%v snapshots found before %v %v  \n", v.ManagedEntity.ExtensibleManagedObject.Self, b.Truncate(time.Second), co)
	//	err = pr.Add(v.ManagedEntity.ExtensibleManagedObject.Self.Value, "Count",co)
	//	if err != nil {
	//		pr.err = err.Error()
	//		_ = pr.Print(start)
	//		return err
	//	}
	//}
	//
	//
	_ = pr.Print(start, txt)

	return nil
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
	var vms []mo.VirtualMachine
	err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"snapshot", "name"}, &vms, f)
	if err != nil {
		return
	}

	pr := NewPrtgData("snapshots")
	tm := newTagMap()
	err = c.tagList(tagIds, tm)
	if err != nil {
		return err
	}
	//vmLoop:
	for _, v := range vms {
		now := time.Now()
		b := now.Add(-age)
		var co int
		if v.Snapshot != nil {
			co, err = snapshotCount(b, v.Snapshot.RootSnapshotList)
			if err != nil {
				return
			}
		}

		if tm.check(v.Self.Value, tagIds) || len(tagIds) == 0 {
			stat := fmt.Sprintf("%v_%v", v.Self.Value, v.Name)
			err = pr.Add(stat, "One", co, lim)
		}

	}

	err = pr.Print(start, txt)
	return

}

func (c *Client) vmMetricS(filter property.Filter) (vm string, m map[string]Prtgitem, err error) {
	m = make(map[string]Prtgitem)

	ctx := context.Background()
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return "", nil, fmt.Errorf("con view 1 %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()

	vmsRefs, err := v.Find(ctx, []string{"VirtualMachine"}, filter) //todo need to fix this  find does not work
	if err != nil {
		return "", nil, fmt.Errorf("%v", err)
	}
	if len(vmsRefs) != 1 {
		return "", nil, fmt.Errorf("filter issue, expected 1 vm and got %v, %v", len(vmsRefs), vmsRefs)
	}

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
		MetricId:   []types.PerfMetricId{}, //{Instance: "*"}
		IntervalId: 300,
	}

	// Query metrics
	sample, err := perfManager.SampleByName(ctx, spec, names, vmsRefs)
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
		//fmt.Printf("%+v\n",metric)
		for _, v := range metric.Value {
			vm = metric.Entity.String()
			counter := counters[v.Name]
			units := counter.UnitInfo.GetElementDescription().Label

			if len(v.Instance) != 0 {

				if debug {
					if len(v.Value) != 0 {
						fmt.Printf("%+v\n", v)
						//fmt.Printf("%T %v\n", v.Value, v.Value) // todo
					}
				}

				st, err := singleStat(v.ValueCSV())
				if err != nil {
					return "", nil, fmt.Errorf("singlestat failed %v", err)
				}
				m[v.Name] = Prtgitem{
					Value: st,
					Unit:  VmMetType(units),
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

func (c *Client) dsMetrics(path string) (ds string, m map[string]Prtgitem, err error) {
	m = make(map[string]Prtgitem)
	ctx := context.Background()
	//dv, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	//if err != nil {
	//	return "", nil, nil
	//}
	//
	//defer dv.Destroy(ctx)
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
	finder := find.NewFinder(c.c, true)
	dc, err := finder.Datacenter(ctx, "*")
	if err != nil {
		return "", nil, nil
	}
	finder.SetDatacenter(object.NewDatacenter(c.c, dc.Reference()))

	finder.SetDatacenter(dc)
	dso, err := finder.Datastore(ctx, "*1")
	if err != nil {
		return "", nil, fmt.Errorf("dsf.datastore %v", err)
	}

	fmt.Printf("dso %+v\n", dso)
	srm := object.NewStorageResourceManager(c.c)
	s, err := srm.QueryDatastorePerformanceSummary(ctx, dso)
	fmt.Println("ds ", s)
	return dso.Name(), nil, err
}

func VmMetType(k string) string {
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
	prtgmap := make(map[string]string)
	prtgmap["KB"] = KiloByte
	prtgmap["MB"] = MegaByte
	prtgmap["GB"] = GigaByte
	prtgmap["TB"] = TeraByte

	prtgmap["num"] = One
	prtgmap["ms"] = TimeResponse
	prtgmap["%"] = Percent
	prtgmap["KBps"] = KiloBit
	prtgmap["KB"] = KiloByte
	prtgmap["MHz"] = CPU
	prtgmap["℃"] = Temperature
	prtgmap["µs"] = Custom
	prtgmap["s"] = Second
	prtgmap["W"] = Custom
	prtgmap["J"] = Custom

	t := prtgmap[k]
	if t == "" {

		return k
	}
	return t
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
	case float32, float64, int8, uint8, int16, uint16, int32, uint32, uint64, int64, uint, int, string, bool:
		if stat == "" {
			return nil, nil
		}
		rtnStat = stat
	case []float64:
		st := stat.([]float64)
		rtnStat = st[0]
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

func printJson(i ...interface{}) {
	for _, v := range i {
		b, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", string(b))
	}
}

func tagCheck(n string, t []string) (found bool) {
	for _, check := range t {
		if n == check {
			return true
		}
	}

	return
}
