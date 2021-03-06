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
	"github.com/juju/fslock"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"io/ioutil"
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
		"cpu.readiness.average", "cpu.usage.average", "sys.uptime.latest",
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
	dsSummaryDefault = []string{
		"datastore.busResets.summation",
		"datastore.commandsAborted.summation",
		"datastore.datastoreIops.average",
		"datastore.datastoreMaxQueueDepth.latest",
		"datastore.datastoreNormalReadLatency.latest",
		"datastore.datastoreNormalWriteLatency.latest",
		"datastore.datastoreReadBytes.latest",
		"datastore.datastoreReadIops.latest",
		"datastore.datastoreReadLoadMetric.latest",
		"datastore.datastoreReadOIO.latest",
		"datastore.datastoreVMObservedLatency.latest",
		"datastore.datastoreWriteBytes.latest",
		"datastore.datastoreWriteIops.latest",
		"datastore.datastoreWriteLoadMetric.latest",
		"datastore.datastoreWriteOIO.latest",
		"datastore.maxTotalLatency.latest",
		"datastore.numberReadAveraged.average",
		"datastore.numberWriteAveraged.average",
		"datastore.read.average",
		"datastore.siocActiveTimePercentage.average",
		"datastore.sizeNormalizedDatastoreLatency.average",
		"datastore.throughput.contention.average",
		"datastore.throughput.usage.average",
		"datastore.totalReadLatency.average",
		"datastore.totalWriteLatency.average",
		"datastore.unmapIOs.summation",
		"datastore.unmapSize.summation",
		"datastore.write.average",
	}
)

func getLock(f string, t time.Duration) (lock *fslock.Lock, err error) {
	lock = fslock.New(f + ".lock")
	for start := time.Now(); time.Since(start) < t; {
		err = lock.TryLock()
		if err == nil {
			return
		}
	}

	return

}

func (c *Client) vmTracker(vm, host string) error {
	type vmTracker struct {
		Host     string
		LastSeen time.Time
	}
	hvmFile := strings.Join([]string{configDir(), "vmTracker.json"}, pathSep)
	lock, err := getLock(hvmFile, 10*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Unlock() }()

	vmBytes, err := ioutil.ReadFile(hvmFile)
	if err != nil {
		fmt.Println(err)
	}
	var vmMap map[string]vmTracker
	err = json.Unmarshal(vmBytes, &vmMap)
	if err != nil {
		fmt.Println(err)
	}

	if vmMap == nil {
		vmMap = make(map[string]vmTracker)
	}

	vmMap[vm] = vmTracker{
		Host:     host,
		LastSeen: time.Now(),
	}

	outJs, err := json.MarshalIndent(vmMap, "", "    ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(hvmFile, outJs, 0644)
	return err
}

func (c *Client) findOne(name, vmwareType string) (moid types.ManagedObjectReference, err error) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 5*time.Second)

	v, err := c.m.CreateContainerView(c.ctx, c.c.ServiceContent.RootFolder, []string{vmwareType}, true)
	if err != nil {
		err = fmt.Errorf("failed to create container %v %v %v", name, vmwareType, err)
		return
	}
	defer func() { _ = v.Destroy(ctx) }()
	switch vmwareType {
	case "HostSystem":
		ol := mo.HostSystem{}
		err = v.RetrieveWithFilter(ctx, []string{"HostSystem"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol.Reference()

	case "VirtualMachine":
		ol := mo.VirtualMachine{}
		err = v.RetrieveWithFilter(ctx, []string{"VirtualMachine"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol.Reference()
	case "Datastore":
		ol := mo.Datastore{}
		err = v.RetrieveWithFilter(ctx, []string{"Datastore"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol.Reference()

	case "VmwareDistributedVirtualSwitch":
		ol := mo.VmwareDistributedVirtualSwitch{}
		err = v.RetrieveWithFilter(ctx, []string{"VmwareDistributedVirtualSwitch"}, []string{"name"}, &ol, property.Filter{"name": name})
		if err != nil {
			err = fmt.Errorf("failed to find %v %v %v", name, vmwareType, err)
			return
		}
		moid = ol.Reference()

	default:
		fmt.Println("findOne() unsupported type")
		return
	}

	return
}

//VMSummary  stats for a VM
func (c *Client) VMSummary(name, moid string, lim *LimitsStruct, age time.Duration, txt bool, sensors []string) error {
	vmSummaryDefault = append(vmSummaryDefault, sensors...)
	start := time.Now()
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)

	if c.m == nil {
		return fmt.Errorf("no manager")
	}
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		return fmt.Errorf("con view 1 %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()

	id := types.ManagedObjectReference{
		Type: "VirtualMachine", Value: moid,
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
		return errCheck(name, id, fmt.Errorf("vm v.properties %v", err))
	}

	elapsed := time.Since(start)

	var co int
	if v0.Snapshot != nil {
		co, err = snapshotCount(time.Now().Add(-age), v0.Snapshot.RootSnapshotList)
		if err != nil {
			return fmt.Errorf("snapshot %v", err)
		}
	}
	pr := newPrtgData(v0.Name)
	pr.moid = id.Value
	_ = pr.add(co, ps.SensorChannel{Channel: fmt.Sprintf("Snapshots Older Than %v", age), Unit: "Custom", CustomUnit: "Found", LimitErrorMsg: lim.ErrMsg, LimitMaxError: lim.MaxErr, LimitMaxWarning: lim.MaxWarn, LimitWarningMsg: lim.WarnMsg})

	gt := ps.SensorChannel{Channel: "guest tools running", Unit: "Custom", ValueLookup: "prtg.standardlookups.exchangedag.yesno.allstatesok"}
	var gtv int
	switch v0.Guest.ToolsRunningStatus {
	case "guestToolsRunning":
		gtv = 1
	default:
		gtv = 0

	}
	_ = pr.add(gtv, gt)

	hs := mo.HostSystem{}
	err = v.Properties(ctx, v0.Runtime.Host.Reference(), []string{"name"}, &hs)
	if err != nil {
		return fmt.Errorf("hostsystem properties failure %v", err)
	}

	for _, v := range v0.Guest.Disk {
		d := v.DiskPath
		if strings.Contains(d, "overlay") {
			continue
		}
		ca := v.Capacity
		free := v.FreeSpace
		one := ca / 100
		perc := free / one
		_ = pr.add(free, ps.SensorChannel{Channel: "free Bytes " + d, Unit: "BytesDisk", VolumeSize: "KiloByte", ShowChart: "0", ShowTable: "0"})
		_ = pr.add(perc, ps.SensorChannel{Channel: "free Space (Percent) " + d, Unit: "Percent", LimitMinWarning: "20", LimitMinError: "10", LimitWarningMsg: "Warning Low Space", LimitErrorMsg: "Critical disk space", LimitMode: "1"})
	}
	if v0.Runtime.PowerState == "poweredOn" {
		pr.text = "OK running on Host " + hs.Name
		err = c.Metrics(v0.Reference(), pr, vmSummaryDefault, 20)
		if err != nil {
			return err
		}
	} else {
		pr.text = fmt.Sprint(v0.Runtime.PowerState)
	}

	err = pr.print(elapsed, txt)
	_ = c.vmTracker(v0.Name, hs.Name)
	return err
}

func errCheck(name string, oid types.ManagedObjectReference, err error) error {
	e := err.Error()
	switch {
	case strings.Contains(e, "has already been deleted"):
		err = fmt.Errorf("object not found %v %v", name, oid)
	default:
	}
	return err
}

//SnapShotsOlderThan tag focused snapshot reporting
func (c *Client) SnapShotsOlderThan(f property.Filter, tagIds []string, lim *LimitsStruct, age time.Duration, txt bool) (err error) {
	start := time.Now()
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)

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
	pr := newPrtgData("snapshots")
	tm := NewTagMap()
	err = c.list(tagIds, tm)
	if err != nil {
		return err
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
				err = pr.add(co, ps.SensorChannel{Channel: stat, Unit: "Custom", CustomUnit: "Found", LimitErrorMsg: lim.ErrMsg, LimitMaxError: lim.MaxErr, LimitMaxWarning: lim.MaxWarn, LimitWarningMsg: lim.WarnMsg})
				if err != nil {
					return
				}
			}
		}(v)

	}

	wg.Wait()
	_ = pr.print(respTime, txt)
	return err

}

//DsSummary stats for a datastore
func (c *Client) DsSummary(name, moid string, lim *LimitsStruct, js bool) (err error) {

	start := time.Now()
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)
	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		return nil
	}

	defer func() { _ = v.Destroy(ctx) }()
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

	ds := mo.Datastore{}
	err = v.Properties(ctx, id, []string{"name", "summary"}, &ds)
	if err != nil {
		return errCheck(name, id, fmt.Errorf("ds v.properties %v", err))
	}
	pr := newPrtgData(ds.Name)
	pr.moid = id.Value
	whole := ds.Summary.Capacity
	free := ds.Summary.FreeSpace
	p1 := whole / 100
	if p1 > 0 {
		freep := free / p1
		provisioned := 100 - freep
		_ = pr.add(freep, ps.SensorChannel{Channel: "Free space (Percent)", Unit: "Percent", DecimalMode: "1", LimitMinWarning: lim.MinWarn, LimitMinError: lim.MinErr, LimitWarningMsg: "Warning Low Space", LimitErrorMsg: "Critical disk space", LimitMode: "1"})
		_ = pr.add(provisioned, ps.SensorChannel{Channel: "Used Space (Percent)", Unit: "Percent", DecimalMode: "1"})

	}
	_ = pr.add(whole, ps.SensorChannel{Channel: "Total capacity", Unit: "BytesDisk", VolumeSize: "KiloByte"})
	_ = pr.add(free, ps.SensorChannel{Channel: "Free Bytes", Unit: "BytesDisk", VolumeSize: "KiloByte", ShowTable: "0", ShowChart: "0"})
	mm := "0"
	if ds.Summary.MaintenanceMode != "normal" {
		mm = "1"
	}

	_ = pr.add(mm, ps.SensorChannel{Channel: "Maintenance Mode", Unit: "Custom", LimitMaxWarning: "1", ValueLookup: "prtg.standardlookups.boolean.statefalseok"})
	_ = pr.add(boolToInt(ds.Summary.Accessible), ps.SensorChannel{Channel: "Accessible", Unit: "Custom", LimitMaxWarning: "1", ValueLookup: "prtg.standardlookups.boolean.statetrueok"})

	err = c.Metrics(ds.Reference(), pr, dsSummaryDefault, 1800)
	if err != nil {
		return err
	}
	_ = pr.print(time.Since(start), js)
	return nil
}

//VdsSummary  stats for a VDS
func (c *Client) VdsSummary(name, moid string, js bool) (err error) {
	start := time.Now()

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)

	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"VmwareDistributedVirtualSwitch"}, true)
	if err != nil {
		return fmt.Errorf("create container %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()

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
		return errCheck(name, id, fmt.Errorf("vds v.properties %v", err))
	}

	elapsed := time.Since(start)
	pr := newPrtgData("VdsSummary")

	_ = pr.add(tfl(vds.OverallStatus), ps.SensorChannel{Channel: "Overall Status", Unit: "Custom", CustomUnit: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health"})
	_ = pr.add(tfl(vds.ConfigStatus), ps.SensorChannel{Channel: "Config Status", Unit: "Custom", CustomUnit: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health"})

	for _, pg := range vds.Portgroup {
		vpg := mo.DistributedVirtualPortgroup{}
		err = v.Properties(ctx, pg, nil, &vpg)
		if err != nil {
			return fmt.Errorf("hs properties %v", err)
		}
		_ = pr.add(tfl(vpg.OverallStatus), ps.SensorChannel{Channel: vpg.Name, Unit: "Custom", CustomUnit: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health"})
	}
	_ = c.Metrics(vds.Reference(), pr, vdsSummaryDefault, 20)
	err = pr.print(elapsed, js)

	return
}

//HostSummary  stats for a host system
func (c *Client) HostSummary(name, moid string, js bool) (err error) {
	start := time.Now()

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)

	v, err := c.m.CreateContainerView(ctx, c.c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		return fmt.Errorf("create container %v", err)
	}
	defer func() { _ = v.Destroy(ctx) }()
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
		return errCheck(name, id, fmt.Errorf("hs v.properties %v", err))
	}

	pr := newPrtgData("HostSummary")

	ps1 := ps.SensorChannel{Channel: "Power state", Unit: "Custom", VolumeSize: "Custom", ValueLookup: "prtg.standardlookups.Google.Gsa.Health", LimitWarningMsg: "Host was put to sleep", LimitErrorMsg: "Host in unknown state, please investigate"}
	switch hs.Runtime.PowerState {
	case "poweredOn":
		_ = pr.add(0, ps1)
	case "poweredOff", "standby":
		_ = pr.add(1, ps1)
		_ = pr.print(time.Since(start), false)
		return
	default:
		_ = pr.add(2, ps1)
		_ = pr.print(time.Since(start), false)
		return
	}

	elapsed := time.Since(start)

	freeMemory := hs.Summary.Hardware.MemorySize - (int64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)
	freeMemoryp := freeMemory / (hs.Summary.Hardware.MemorySize / 100)

	_ = pr.add(freeMemory, ps.SensorChannel{Channel: "Memory Free", Unit: "BytesMemory"})
	_ = pr.add(freeMemoryp, ps.SensorChannel{Channel: "Memory Free (Percent)", Unit: "Percent", DecimalMode: "1"})

	totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
	freeCPU := totalCPU - int64(hs.Summary.QuickStats.OverallCpuUsage)
	usedCPUP := int64(hs.Summary.QuickStats.OverallCpuUsage) / (totalCPU / 100)
	freeCPUP := freeCPU / (totalCPU / 100)
	_ = pr.add(freeCPUP, ps.SensorChannel{Channel: "CPU Free", Unit: "Percent", DecimalMode: "1"})
	_ = pr.add(usedCPUP, ps.SensorChannel{Channel: "CPU Used", Unit: "Percent", DecimalMode: "1"})

	_ = pr.add(totalCPU-freeCPU, ps.SensorChannel{Channel: "CPU Used MHz", Unit: "Custom", VolumeSize: "One", CustomUnit: "MHz"})
	_ = pr.add(freeCPU, ps.SensorChannel{Channel: "CPU Free MHz", Unit: "Custom", VolumeSize: "One", CustomUnit: "MHz"})
	_ = pr.add(totalCPU, ps.SensorChannel{Channel: "CPU Capacity MHz", Unit: "Custom", VolumeSize: "One", CustomUnit: "MHz"})

	_ = pr.add(boolToInt(hs.Runtime.InMaintenanceMode), ps.SensorChannel{Channel: "Maintenance Mode", Unit: "Custom", VolumeSize: "Custom", ValueLookup: "prtg.standardlookups.boolean.statefalseok"})
	_ = pr.add(triggeredAlarms(hs.TriggeredAlarmState), ps.SensorChannel{Channel: "Triggered Alarms", Unit: "Count", LimitMaxWarning: "1", LimitWarningMsg: "triggered alarms present"})

	pr.text = fmt.Sprint(hs.Runtime.PowerState)
	diskPathDetails := hs.Config.StorageDevice.MultipathInfo
	var triggered bool
	for _, lun := range diskPathDetails.Lun {
		for _, path := range lun.Path {
			if !*path.IsWorkingPath {
				triggered = true
				pr.text = "Path failure " + path.Name
			}
		}
	}
	_ = pr.add(boolToInt(triggered), ps.SensorChannel{Channel: "storage_path_error", Unit: "Custom", VolumeSize: "Custom", ValueLookup: "prtg.standardlookups.boolean.statefalseok", LimitErrorMsg: "check storage paths"})
	err = c.Metrics(id, pr, hsSummaryDefault, 20)
	if err != nil {
		return
	}
	_ = pr.print(elapsed, js)
	return
}

func (c *Client) getMaxQueryMetrics(ctx context.Context) (int, error) {

	om := object.NewOptionManager(c.c, *c.c.ServiceContent.Setting)
	res, err := om.Query(ctx, "config.vpxd.stats.maxQueryMetrics")
	if err == nil {
		if len(res) > 0 {
			if s, ok := res[0].GetOptionValue().Value.(string); ok {
				v, err := strconv.Atoi(s)
				if err != nil {
					return 0, err
				}

				if v == -1 {
					// Whatever the server says, we never ask for more metrics than this.
					return 10000, nil
				}
				return v, nil
			}
		}
		// Fall through version-based inference if value isn't usable

	}

	// No usable maxQueryMetrics setting. Infer based on version
	ver := c.c.ServiceContent.About.Version
	parts := strings.Split(ver, ".")
	if len(parts) < 2 {
		return 64, nil
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	if major < 6 || major == 6 && parts[1] == "0" {
		return 64, nil
	}
	return 256, nil
}

//Metrics returns metrics for a given object
func (c *Client) Metrics(mor types.ManagedObjectReference, pr *prtgData, str []string, interval int32) (err error) {

	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, 30*time.Second)

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
		MetricId:   []types.PerfMetricId{},
		IntervalId: interval,
	}

	maxQuery, err := c.getMaxQueryMetrics(ctx)
	if err != nil {
		return fmt.Errorf("getMaxQueryMetrics %v", err)
	}
	psum, err := perfManager.ProviderSummary(ctx, mor)
	if err != nil {
		return fmt.Errorf("object not found %v", mor)
	}
	if !psum.CurrentSupported {
		fmt.Printf("%v performance metrics not available\n", mor)
		return
	}

	// Query metrics
	sample, err := perfManager.SampleByName(ctx, spec, names, []types.ManagedObjectReference{mor})
	if (err != nil) || len(sample) == 0 {
		return fmt.Errorf("could not find sample data for %v, err: %v", mor, err)
	}

	// dont run query if it will exceed current limit
	if len(sample) > maxQuery {
		return fmt.Errorf("query metrics level too low, needed %v  max setting %v", len(sample), maxQuery)
	}

	result, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return fmt.Errorf("metrics %v", err)
	}

	res := result[0]

	//Read result
	sort.Slice(res.Value, func(i, j int) bool {
		return res.Value[i].Name < res.Value[j].Name
	})

	for _, v := range res.Value {
		if (len(v.Value) == 0) || v.Value[0] == -1 {
			continue
		}

		var hide bool
		counter := counters[v.Name]
		instance := v.Instance
		if inStringSlice(v.Name, str) {
			if instance != "" {
				// special handling of metric names using instance data
				switch mor.Type {
				case "VmwareDistributedVirtualSwitch":
					// drop net.throughput prefix
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
			//if instance == "" {
			//	instance = "-"
			//}

			units := counter.UnitInfo.GetElementDescription().Label

			// special handling for int's
			fixedPointFloat := float64(v.Value[0]) / 100
			if strings.Contains(v.Name, "power") || strings.Contains(v.Name, "sys.uptime") {
				fixedPointFloat = float64(v.Value[0])
			}

			// get PRTG version of vmware metric, eg type % == Percent
			u, s, cu := metType(units, counter.GroupInfo.GetElementDescription().Key)

			// allow hiding of verbose channels
			if !hide {
				_ = pr.add(fixedPointFloat, ps.SensorChannel{Channel: v.Name, Unit: u, VolumeSize: s, CustomUnit: cu}) //, DecimalMode: decMode})

			} else {
				_ = pr.add(fixedPointFloat, ps.SensorChannel{Channel: v.Name, Unit: u, VolumeSize: s, CustomUnit: cu, ShowChart: "0", ShowTable: "0"})

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

func metType(u, s string) (unit, size, customUnit string) {
	//noinspection GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst,GoUnusedConst
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
		}
	case "MHz":
		unit = Custom
		size = One
		customUnit = "MHz"
	case "℃":
		size = Temperature
	case "µs":
		size = Custom
	case "s":
		unit = "TimeSeconds"
	case "W":
		size = Custom
		customUnit = "Watt"
	default:
		size = u
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

func printJSON(txt bool, i ...interface{}) {
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

//func tagCheck(n string, t []string) (found bool) {
//	for _, check := range t {
//		if n == check {
//			return true
//		}
//	}
//
//	return
//}

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

	case "yellow":
		i = 1
	case "red":
		i = 2
	default:
		fmt.Println("tfl", ic)
		i = 9
	}
	return i
}
