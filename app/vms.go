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
	"context"
	"encoding/json"
	"fmt"
	ps "github.com/PRTG/go-prtg-sensor-api"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	vmSummaryDefault = []string{
		"disk.read.average", "disk.write.average", "disk.usage.average",
		"cpu.readiness.average", "cpu.usage.average",
		"mem.active.average", "mem.consumed.average", "mem.usage.average",
		"net.bytesRx.average", "net.bytesTx.average", "net.usage.average",
		"datastore.datastoreNormalReadLatency.latest", "datastore.datastoreNormalWriteLatency.latest",
		"datastore.datastoreReadIops.latest", "datastore.datastoreWriteIops.latest",
	}
	hsSummaryDefault = []string{"cpu.latency.average", "cpu.readiness.average", "cpu.usage.average",
		"disk.read.average", "disk.usage.average", "disk.write.average",
		"mem.active.average", "mem.consumed.average", "mem.llSwapUsed.average", "mem.compressionRate.average",
		"net.received.average", "net.transmitted.average", "net.usage.average",
		"power.power.average",
	}
	vdsSummaryDefault = []string{
		"net.throughput.droppedRx.average", "net.throughput.droppedTx.average",
		"net.throughput.pktsRx.average", "net.throughput.pktsTx.average",
		"net.throughput.pktsRxBroadcast.average", "net.throughput.pktsTxBroadcast.average",
		"net.throughput.pktsRxMulticast.average", "net.throughput.pktsTxMulticast.average",
		"net.throughput.vds.droppedRx.average", "net.throughput.vds.droppedTx.average",

		"net.throughput.vds.pktsRx.average", "net.throughput.vds.pktsRx.average",
		"net.throughput.vds.pktsRxBcast.average", "net.throughput.vds.pktsTx.average",
		"net.throughput.vds.pktsRxMcast.average", "net.throughput.vds.pktsTxMcast.average",
		"net.throughput.vds.pktsRxBcast.average", "net.throughput.vds.pktsTxBcast.average",
	}
)

func (c *Client) findOne(name, vmwareType string) (moid types.ManagedObjectReference, err error) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	v, err := c.m.CreateContainerView(c.ctx, c.c.ServiceContent.RootFolder, []string{vmwareType}, true)
	if err != nil {
		err = fmt.Errorf("failed to create container %v %v %v", name, vmwareType, err)
		return
	}
	defer v.Destroy(ctx)

	switch vmwareType {
	case "HostSystem":
		ol := []mo.HostSystem{}
		err = v.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol[0].Reference()

	case "VirtualMachine":
		ol := []mo.VirtualMachine{}
		err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol[0].Reference()
	case "Datastore":
		ol := []mo.Datastore{}
		err = v.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol[0].Reference()

	case "VmwareDistributedVirtualSwitch":
		ol := []mo.VmwareDistributedVirtualSwitch{}
		err = v.RetrieveWithFilter(ctx, []string{"VmwareDistributedVirtualSwitch"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol[0].Reference()

	default:
		fmt.Println("findOne() unsupported type")
		return
	}

	return
}

//VmSummary
func (c *Client) VmSummary(name, moid string, lim *LimitsStruct, age time.Duration, txt bool, sensors []string) error {
	vmSummaryDefault = append(vmSummaryDefault, sensors...)
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
	//vms := make([]mo.VirtualMachine, 0, 100)
	id := types.ManagedObjectReference{
		"VirtualMachine", moid,
	}

	v0 := mo.VirtualMachine{}

	if moid == "" {
		id, err = c.findOne(name, "VirtualMachine")
		if err != nil {
			return fmt.Errorf("c.findOne %v", err)
		}
	}
	err = v.Properties(ctx, id, []string{"name", "summary", "snapshot", "guest", "runtime"}, &v0)
	if err != nil {
		return fmt.Errorf("v.properties %v", err)
	}
	//if len(vms) != 1 {
	//
	//	type vmFailList struct {
	//		name, moid string
	//	}
	//	out := make([]vmFailList, 0, 10)
	//	for _, v := range vms {
	//		out = append(out, vmFailList{v.Name, v.Self.Value})
	//
	//	}
	//
	//	return fmt.Errorf("expected a single vm, got %+v", out)
	//}

	//	v0 := vms[0]
	//printJson(false,item)
	//vm, mets, err := c.vmMetricS(v0.Reference())
	//if err != nil {
	//	return fmt.Errorf("metrics %v", err)
	//}
	elapsed := time.Since(start)

	var co int
	if v0.Snapshot != nil {
		co, err = snapshotCount(time.Now().Add(-age), v0.Snapshot.RootSnapshotList)
		if err != nil {
			return fmt.Errorf("snapshot %v", err)
		}
	}
	pr := NewPrtgData(v0.Name)
	pr.moid = id.Value
	_ = pr.Add(co, ps.SensorChannel{Channel: fmt.Sprintf("Snapshots Older Than %v", age), Unit: "Custom", CustomUnit: "Found", LimitErrorMsg: lim.ErrMsg, LimitMaxError: lim.MaxErr, LimitMaxWarning: lim.MaxWarn, LimitWarningMsg: lim.WarnMsg})

	//	guestLimits := &LimitsStruct{
	//		MinErr: 0.5,
	//		ErrMsg: "tools not running",
	//	}

	gt := ps.SensorChannel{Channel: "guest tools running", Unit: "Custom", ValueLookup: "prtg.standardlookups.exchangedag.yesno.allstatesok"}
	var gtv int
	switch v0.Guest.ToolsRunningStatus {
	case "guestToolsRunning":
		gtv = 1
	default:
		gtv = 0

	}
	_ = pr.Add(gtv, gt)

	hs := mo.HostSystem{}
	err = v.Properties(ctx, v0.Runtime.Host.Reference(), []string{"name"}, &hs)
	if err != nil {
		return fmt.Errorf("hostsystem properties failure %v", err)
	}

	pr.text = "on host " + hs.Name
	err = c.MetricS(v0.Reference(), pr, vmSummaryDefault, 20)
	if err != nil {
		return err
	}
	//for _, v := range mets {
	//	if inStringSlice(v.Channel, vmSummaryDefault) {
	//		st, err := singleStat(v.Value)
	//		if err != nil {
	//			return err
	//		}
	//
	//		if st != "" {
	//			err = pr.Add(st, v)
	//			if err != nil {
	//				return err
	//			}
	//		}
	//	}
	//}
	for _, v := range v0.Guest.Disk {
		d := v.DiskPath
		ca := v.Capacity
		free := v.FreeSpace
		one := ca / 100
		perc := free / one
		_ = pr.Add(free/1000, ps.SensorChannel{Channel: "free Bytes " + d, Unit: "BytesDisk", VolumeSize: "KiloByte", ShowChart: "0", ShowTable: "0"})
		_ = pr.Add(perc, ps.SensorChannel{Channel: "free Space (Percent) " + d, Unit: "Percent", LimitMinWarning: "20", LimitMinError: "10", LimitWarningMsg: "Warning Low Space", LimitErrorMsg: "Critical disk space"})
	}
	err = pr.Print(elapsed, txt)

	return err
}

//SnapShotsOlderThan
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
		return fmt.Errorf("retrieve issue %v", err)
	}

	// retrieve tags and object associations
	pr := NewPrtgData("snapshots")
	tm := NewTagMap()
	err = c.list(tagIds, tm)
	if err != nil {
		return err
	}

	if (err != nil) && !strings.Contains(err.Error(), "404") {
		if err != nil {
			return fmt.Errorf("404 issue %v", err)
		}

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
				stat := fmt.Sprintf("%v", v.Name)
				err = pr.Add(co, ps.SensorChannel{Channel: stat, Unit: "Custom", CustomUnit: "Found", LimitErrorMsg: lim.ErrMsg, LimitMaxError: lim.MaxErr, LimitMaxWarning: lim.MaxWarn, LimitWarningMsg: lim.WarnMsg})
				if err != nil {
					return
				}
			}
		}(v)

	}

	wg.Wait()
	_ = pr.Print(respTime, txt)
	return err

}

//DsSummary
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
		Value: moid,
	}

	if moid == "" {
		id, err = c.findOne(name, id.Type)
		if err != nil {
			return fmt.Errorf("c.findOne %v", err)
		}

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
	if p1 > 0 {
		freep := free / p1
		provisioned := 100 - freep
		_ = pr.Add(freep, ps.SensorChannel{Channel: "Free space (Percent)", Unit: "Percent", DecimalMode: "1", LimitMinWarning: lim.MinWarn, LimitMinError: lim.MinErr, LimitWarningMsg: "Warning Low Space", LimitErrorMsg: "Critical disk space"})
		_ = pr.Add(provisioned, ps.SensorChannel{Channel: "Used Space (Percent)", Unit: "Percent", DecimalMode: "1"})

	}
	_ = pr.Add(whole, ps.SensorChannel{Channel: "Total capacity", Unit: "BytesDisk", VolumeSize: "KiloByte"})
	_ = pr.Add(free, ps.SensorChannel{Channel: "Free Bytes", Unit: "BytesDisk", VolumeSize: "KiloByte", ShowTable: "0", ShowChart: "0"})
	mm := "0"
	if v0.Summary.MaintenanceMode != "normal" {
		mm = "1"
	}

	_ = pr.Add(mm, ps.SensorChannel{Channel: "Maintenance Mode", Unit: "Custom", LimitMaxWarning: "1", ValueLookup: "prtg.standardlookups.boolean.statefalseok"})
	err = c.MetricS(v0.Reference(), pr, vdsSummaryDefault, 1800)
	if err != nil {
		return err
	}
	err = pr.Print(time.Since(start), js)
	return nil
}

//VdsSummary
func (c *Client) VdsSummary(name, moid string, lim *LimitsStruct, js bool) (err error) {
	start := time.Now()

	ctx := context.Background()
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return fmt.Errorf("create container %v", err)
	}
	defer v.Destroy(ctx)
	id := types.ManagedObjectReference{
		Type:  "VmwareDistributedVirtualSwitch",
		Value: moid,
	}
	if moid == "" {
		id, err = c.findOne(name, "VmwareDistributedVirtualSwitch")
		if err != nil {
			return
		}
	}
	vds := mo.VmwareDistributedVirtualSwitch{}
	err = v.Properties(ctx, id, nil, &vds)
	if err != nil {
		return fmt.Errorf("vds properties %v", err)
	}

	elapsed := time.Since(start)
	vd := vds
	pr := NewPrtgData("VdsSummary")

	pr.Add(tfl(vd.OverallStatus), ps.SensorChannel{Channel: "Overall Status", Unit: "Custom", CustomUnit: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health"})
	pr.Add(tfl(vd.ConfigStatus), ps.SensorChannel{Channel: "Config Status", Unit: "Custom", CustomUnit: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health"})

	for _, pg := range vd.Portgroup {
		vpg := mo.DistributedVirtualPortgroup{}
		err = v.Properties(ctx, pg, nil, &vpg)
		if err != nil {
			return fmt.Errorf("hs properties %v", err)
		}
		pr.Add(tfl(vpg.OverallStatus), ps.SensorChannel{Channel: vpg.Name, Unit: "Custom", CustomUnit: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health"})
	}
	err = c.MetricS(vd.Reference(), pr, vdsSummaryDefault, 20)
	if err != nil {
		return err
	}
	_ = pr.Print(elapsed, js)
	return
}

//HostSummary
func (c *Client) HostSummary(name, moid string, lim *LimitsStruct, js bool) (err error) {
	start := time.Now()

	ctx := context.Background()
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return fmt.Errorf("create container %v", err)
	}
	defer v.Destroy(ctx)
	id := types.ManagedObjectReference{
		Type:  "HostSystem",
		Value: moid,
	}

	if moid == "" {
		id, err = c.findOne(name, "HostSystem")
		if err != nil {
			return
		}
	}
	hs := mo.HostSystem{}
	err = v.Properties(ctx, id, nil, &hs)
	if err != nil {
		return fmt.Errorf("hs properties %v", err)
	}

	pr := NewPrtgData("HostSummary")

	ps1 := ps.SensorChannel{Channel: "Power state", Unit: "Custom", VolumeSize: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health", LimitWarningMsg: "host was put to sleep", LimitErrorMsg: "host in unknown state, please investigate"}
	switch hs.Runtime.PowerState {
	case "poweredOn":
		_ = pr.Add(0, ps1)
	case "poweredOff", "standby":
		_ = pr.Add(1, ps1)
		pr.Print(time.Since(start), false)
		return
	case "unknown":
		_ = pr.Add(2, ps1)
		pr.Print(time.Since(start), false)
		return
	default:
		printJson(false, hs.Runtime.PowerState)
	}

	err = c.MetricS(id, pr, hsSummaryDefault, 20)
	if err != nil {
		return
	}
	elapsed := time.Since(start)

	freeMemory := int64(hs.Summary.Hardware.MemorySize) - (int64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)
	freeMemoryp := freeMemory / (int64(hs.Summary.Hardware.MemorySize) / 100)
	//
	//err = pr.Add(memUsed, ps.SensorChannel{Channel: "Memory Total", Unit: "BytesMemory"})
	_ = pr.Add(freeMemory, ps.SensorChannel{Channel: "Memory Free", Unit: "BytesMemory"})
	_ = pr.Add(freeMemoryp, ps.SensorChannel{Channel: "Memory Free (Percent)", Unit: "Percent", DecimalMode: "1"})

	totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
	freeCPU := int64(totalCPU) - int64(hs.Summary.QuickStats.OverallCpuUsage)
	usedCPUP := int64(hs.Summary.QuickStats.OverallCpuUsage) / (totalCPU / 100)
	freeCPUP := freeCPU / (totalCPU / 100)
	_ = pr.Add(freeCPUP, ps.SensorChannel{Channel: "CPU Free", Unit: "Percent", DecimalMode: "1"})
	_ = pr.Add(usedCPUP, ps.SensorChannel{Channel: "CPU Used", Unit: "Percent", DecimalMode: "1"})

	_ = pr.Add(totalCPU-freeCPU, ps.SensorChannel{Channel: "CPU Used MHz", Unit: "Custom", VolumeSize: "One", CustomUnit: "MHz"})
	_ = pr.Add(freeCPU, ps.SensorChannel{Channel: "CPU Free MHz", Unit: "Custom", VolumeSize: "One", CustomUnit: "MHz"})
	_ = pr.Add(totalCPU, ps.SensorChannel{Channel: "CPU Capacity MHz", Unit: "Custom", VolumeSize: "One", CustomUnit: "MHz"})

	_ = pr.Add(boolToInt(hs.Runtime.InMaintenanceMode), ps.SensorChannel{Channel: "Maintenance Mode", Unit: "Custom", VolumeSize: "Custom", ValueLookup: "prtg.standardlookups.boolean.statefalseok"})
	_ = pr.Add(triggeredAlarms(hs.TriggeredAlarmState), ps.SensorChannel{Channel: "Triggered Alarms", Unit: "Count", LimitMaxWarning: "1", LimitWarningMsg: "triggered alarms present"})
	if err != nil {
		return err
	}
	_ = pr.Print(elapsed, js)
	return
}

//MetricS
func (c *Client) MetricS(mor types.ManagedObjectReference, pr *PrtgData, str []string, interval int32) (err error) {
	// get object quickstats
	if c.m == nil {
		return fmt.Errorf("Metrics() no client")
	}
	ctx := context.Background()
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{mor.Type}, true)
	if err != nil {
		return fmt.Errorf("con view 1 %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()

	perfManager := performance.NewManager(c.c)

	// Retrieve counters
	counters, err := perfManager.CounterInfoByName(ctx)
	if err != nil {
		return fmt.Errorf("perfmanager %v", err)
	}

	var names []string
	for name := range counters {
		names = append(names, name)
	}

	// Check PerfQuerySpec
	spec := types.PerfQuerySpec{
		MaxSample:  1,
		MetricId:   []types.PerfMetricId{{Instance: "*"}},
		IntervalId: interval,
	}
	psum, err := perfManager.ProviderSummary(ctx, mor)
	if err != nil {
		return err
	}
	if psum.CurrentSupported {

		// Query metrics
		sample, err := perfManager.SampleByName(ctx, spec, names, []types.ManagedObjectReference{mor})
		if err != nil {
			return fmt.Errorf("could not find sample data for %v, err: %v\n", mor, err)
		}

		result, err := perfManager.ToMetricSeries(ctx, sample)
		if err != nil {
			return fmt.Errorf("ToMetricSeries %v", err)
		}

		if len(result) == 0 {

			return fmt.Errorf("no performance data is available for %v", mor.Value)
		}

		//Read result
		for _, metric := range result {
			sort.Slice(metric.Value, func(i, j int) bool {
				return metric.Value[i].Name < metric.Value[j].Name
			})

			for _, v := range metric.Value {
				var hide bool
				counter := counters[v.Name]
				instance := v.Instance
				if inStringSlice(v.Name, str) {
					if instance != "" {
						// special handling of metric names

						switch mor.Type {
						case "VmwareDistributedVirtualSwitch":
							// drop net.throughput
							v.Name = v.Name[15:]
							fields := strings.Fields(instance)
							switch len(fields) {
							case 1:
								v.Name = fmt.Sprintf("%v.%v", v.Name, fields[0])
							case 2:
								hide = true
								v.Name = fmt.Sprintf("%v Port %v %v", fields[0], fields[1], v.Name)
							default:
								v.Name = fmt.Sprintf("%v %v", fields, v.Name)
							}

						default:
							continue
						}

					}
					if len(v.Value) != 0 {
						units := counter.UnitInfo.GetElementDescription().Label
						if v.Value[0] == -1 {
							fmt.Println(v.Name, v.Value)
							continue
						}

						// special handling for power as this returns an int
						var fixedPointFloat float64
						if strings.Contains(v.Name, "power") {
							fixedPointFloat = float64(v.Value[0])

						} else {
							fixedPointFloat = float64(v.Value[0]) / 100
						}

						// force decimal places for percentages
						var decMode string
						if units == "%" {
							decMode = "1"

						}

						// get PRTG version of vmware metric, eg type % == Percent
						u, s, cu := VmMetType(units, counter.GroupInfo.GetElementDescription().Key)

						// allow hiding of verbose channels
						if !hide {
							err = pr.Add(fixedPointFloat, ps.SensorChannel{Channel: v.Name, Unit: u, VolumeSize: s, CustomUnit: cu, DecimalMode: decMode})
							if err != nil {
								return err
							}
						} else {
							err = pr.Add(fixedPointFloat, ps.SensorChannel{Channel: v.Name, Unit: u, VolumeSize: s, CustomUnit: cu, DecimalMode: decMode, ShowChart: "0", ShowTable: "0"})
							if err != nil {
								return err
							}
						}
					}
				}

			}
		}
	}
	return
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

func VmMetType(u, s string) (unit, size, customUnit string) {
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
		case "net":
			unit = "SpeedNet"
		case "disk", "virtualDisk", "datastore", "storageAdapter", "mem", "hbr", "storagePath":
			unit = "SpeedDisk"

		default:
			unit = "Custom"
			customUnit = s
			printJson(false, "missed KBps type", s)

		}
	case "MHz":
		unit = Custom
		size = One
		customUnit = "MHz"
	case "℃":
		size = Temperature
	case "µs":
		size = Custom
	case "W":
		size = Custom
		customUnit = "Watt"
	default:
		size = u
		printJson(false, u)

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
		return nil, fmt.Errorf("type of %v %T is not supported for \n %v\n", t, stat, stat)
	}

	return rtnStat, nil
}

func singleStat2(stat string) (string, error) {
	if stat == "" {
		return "", nil
	}
	fl, err := strconv.ParseFloat(stat, 64)
	if err != nil {
		return "", fmt.Errorf("cant parse this  %v", err)
	}
	fl = float64(int(fl*100) / 100)
	rtnStat := fmt.Sprintf("%v", fl)
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

func triggeredAlarms(s []types.AlarmState) (rntInt int) {
	if s != nil {
		rntInt = len(s)
	}
	return
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func tfl(ic interface{}) int {
	c := fmt.Sprintf("%v", ic)
	var i int
	switch c {
	case "green":

	case "yelllow":
		i = 1
	case "red":
		i = 2
	default:
		fmt.Println("tfl", ic)
		i = 9
	}
	return i
}
