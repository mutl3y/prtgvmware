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
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"
)

var u, _ = url.Parse(os.Getenv("vmurl"))
var user = os.Getenv("vmuser")
var passwd = os.Getenv("vmpass")
var timestamp = time.Now().Truncate(time.Hour)

func TestClient_vmSummary(t *testing.T) {
	type args struct {
		searchName, searchMoid string
		usr, pw                string
		txt                    bool
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"", args{"unknown", "", user, passwd, false}, true},
		{"", args{"mh-cache", "", user, passwd, false}, false},
	}
	//	debug = true
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, tt.args.usr, tt.args.pw, true)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer func() { _ = c.Logout() }()
			lim := &LimitsStruct{}
			err = c.VMSummary(tt.args.searchName, tt.args.searchMoid, lim, time.Hour, tt.args.txt, []string{"cpu.ready.summation"})
			if (err != nil) && !tt.wantErr {
				t.Fatal(err)
			}
		})
	}
}

func Test_snapshotCount(t *testing.T) {

	type args struct {
		snp []types.VirtualMachineSnapshotTree
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{{
		name: "",
		args: args{snp: []types.VirtualMachineSnapshotTree{
			{
				Snapshot: types.ManagedObjectReference{
					Type:  "VirtualMachineSnapshot",
					Value: "parent",
				},
				Vm: types.ManagedObjectReference{
					Type:  "VirtualMachine",
					Value: "vm-26",
				},
				Name:           "test1",
				Description:    "test-snaphot",
				Id:             1,
				CreateTime:     timestamp,
				State:          "poweredOn",
				Quiesced:       true,
				BackupManifest: "",
				ChildSnapshotList: []types.VirtualMachineSnapshotTree{
					{
						Snapshot: types.ManagedObjectReference{
							Type:  "VirtualMachineSnapshot",
							Value: "child 0",
						},
						Vm: types.ManagedObjectReference{
							Type:  "VirtualMachine",
							Value: "vm-26",
						},
						Name:              "test2",
						Description:       "test-sub-hot",
						Id:                2,
						CreateTime:        timestamp,
						State:             "poweredOn",
						Quiesced:          true,
						BackupManifest:    "",
						ChildSnapshotList: nil,
						ReplaySupported:   nil,
					}, {
						Snapshot: types.ManagedObjectReference{
							Type:  "VirtualMachineSnapshot",
							Value: "child 1",
						},
						Vm: types.ManagedObjectReference{
							Type:  "VirtualMachine",
							Value: "vm-26",
						},
						Name:           "test3",
						Description:    "test-sub-hot",
						Id:             3,
						CreateTime:     timestamp,
						State:          "poweredOn",
						Quiesced:       true,
						BackupManifest: "",
						ChildSnapshotList: []types.VirtualMachineSnapshotTree{
							{
								Snapshot: types.ManagedObjectReference{
									Type:  "VirtualMachineSnapshot",
									Value: "child 1 child",
								},
								Vm: types.ManagedObjectReference{
									Type:  "VirtualMachine",
									Value: "vm-26",
								},
								Name:              "test4",
								Description:       "test-sub-hot",
								Id:                4,
								CreateTime:        timestamp,
								State:             "poweredOn",
								Quiesced:          true,
								BackupManifest:    "",
								ChildSnapshotList: nil,
								ReplaySupported:   nil,
							}},
						ReplaySupported: nil,
					}},
				ReplaySupported: nil,
			}}},
		want:    4,
		wantErr: false,
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := snapshotCount(time.Now(), tt.args.snp)
			if (err != nil) != tt.wantErr {
				t.Errorf("snapshotCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("snapshotCount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_snapshotCount(b *testing.B) {
	ti := time.Now().Truncate(time.Second)
	type args struct {
		snp []types.VirtualMachineSnapshotTree
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{{
		name: "",
		args: args{snp: []types.VirtualMachineSnapshotTree{
			{
				Snapshot: types.ManagedObjectReference{
					Type:  "VirtualMachineSnapshot",
					Value: "parent",
				},
				Vm: types.ManagedObjectReference{
					Type:  "VirtualMachine",
					Value: "vm-26",
				},
				Name:           "test1",
				Description:    "test-snaphot",
				Id:             1,
				CreateTime:     ti,
				State:          "poweredOn",
				Quiesced:       true,
				BackupManifest: "",
				ChildSnapshotList: []types.VirtualMachineSnapshotTree{
					{
						Snapshot: types.ManagedObjectReference{
							Type:  "VirtualMachineSnapshot",
							Value: "child 0",
						},
						Vm: types.ManagedObjectReference{
							Type:  "VirtualMachine",
							Value: "vm-26",
						},
						Name:              "test2",
						Description:       "test-sub-hot",
						Id:                2,
						CreateTime:        ti,
						State:             "poweredOn",
						Quiesced:          true,
						BackupManifest:    "",
						ChildSnapshotList: nil,
						ReplaySupported:   nil,
					}, {
						Snapshot: types.ManagedObjectReference{
							Type:  "VirtualMachineSnapshot",
							Value: "child 1",
						},
						Vm: types.ManagedObjectReference{
							Type:  "VirtualMachine",
							Value: "vm-26",
						},
						Name:           "test3",
						Description:    "test-sub-hot",
						Id:             3,
						CreateTime:     ti,
						State:          "poweredOn",
						Quiesced:       true,
						BackupManifest: "",
						ChildSnapshotList: []types.VirtualMachineSnapshotTree{
							{
								Snapshot: types.ManagedObjectReference{
									Type:  "VirtualMachineSnapshot",
									Value: "child 1 child",
								},
								Vm: types.ManagedObjectReference{
									Type:  "VirtualMachine",
									Value: "vm-26",
								},
								Name:              "test4",
								Description:       "test-sub-hot",
								Id:                4,
								CreateTime:        ti,
								State:             "poweredOn",
								Quiesced:          true,
								BackupManifest:    "",
								ChildSnapshotList: nil,
								ReplaySupported:   nil,
							}},
						ReplaySupported: nil,
					}},
				ReplaySupported: nil,
			}}},
		want:    4,
		wantErr: false,
	},
	}
	for _, tt := range tests {
		for n := 0; n < b.N; n++ {
			got, err := snapshotCount(time.Now(), tt.args.snp)
			if err != nil {
				b.Fatalf("failed %v", err)
			}
			if got != tt.want {
				b.Fatalf("value mismatch got %v  wanted %v", got, tt.want)
			}
		}
	}
}

func TestSnapShotsOlder(t *testing.T) {
	type args struct {
		searchType, searchItem string
		usr, pw                string
		tag                    []string
		txt                    bool
	}
	tests := []struct {
		name    string
		ur      *url.URL
		args    args
		wantErr bool
	}{
		{"5", u, args{"name", "mh-cache", user, passwd, []string{"windows", "PRTG"}, false}, false},
		{"6", u, args{"name", "mh-cache", user, passwd, []string{"windowsx"}, false}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewClient(tt.ur, tt.args.usr, tt.args.pw, true)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer func() { _ = c.Logout() }()

			f := property.Filter{tt.args.searchType: "*" + tt.args.searchItem}
			lim := &LimitsStruct{}

			err = c.SnapShotsOlderThan(f, tt.args.tag, lim, time.Second, tt.args.txt)
			if (err != nil) && !tt.wantErr {
				t.Errorf("failed %v", err)
			}
			if !c.Cached {
				_ = c.Logout()
			}
		})

	}
}
func Benchmark_SnapShotsOlder(b *testing.B) {
	type args struct {
		searchType, searchItem string
		usr, pw                string
		tag                    []string
		txt                    bool
	}

	tests := []struct {
		name    string
		ur      *url.URL
		args    args
		wantErr bool
	}{
		{"5", u, args{"name", "mh-cache", "prtg@heynes.local", ".l3tm31n", []string{"windows", "PRTG"}, false}, false},
	}
	for _, tt := range tests {
		wg := sync.WaitGroup{}
		for n := 0; n < 1000; n++ {
			wg.Add(1)
			c, err := NewClient(tt.ur, tt.args.usr, tt.args.pw, true)
			if err != nil {
				b.Fatalf("%+v", err)
			}
			f := property.Filter{tt.args.searchType: "*" + tt.args.searchItem}
			lim := &LimitsStruct{}

			go func() {
				defer wg.Done()
				err = c.SnapShotsOlderThan(f, tt.args.tag, lim, time.Second, tt.args.txt)
				if (err != nil) && !tt.wantErr {
					b.Errorf("failed %v", err)
				}
			}()

		}
		wg.Wait()
	}
}

//func TestClient_vmMstrics(t *testing.T) {
//	tests := []struct {
//		name    string
//		prop    property.Filter
//		wantErr bool
//	}{
//		{"", property.Filter{"name": "mh-cache"}, false},
//		//{"", property.Filter{"self": "*"}, true},
//		//{"", property.Filter{"self": "*"}, true},
//		//{"", property.Filter{"name": "*2"}, true},
//		//{"", nil, true},
//	}
//	u, err := url.Parse("https://192.168.0.201/sdk")
//	if err != nil {
//		t.Fatalf("failed to parse url")
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", timeout)
//			if err != nil {
//				t.Errorf("failed %v", err)
//			}
//
//			_, mets, err := c.vmMetricS(types.ManagedObjectReference{
//				Type:  "VirtualMachine",
//				Value: "vm-16",
//			})
//			if (err != nil) != tt.wantErr {
//				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
//			}
//
//			printJSON(false, mets)
//		})
//	}
//}

func TestClient_DsSummarys(t *testing.T) {

	tests := []struct {
		name    string
		na      string
		moid    string
		wantErr bool
	}{
		{"fail", "", "datastore-1", true},
		{"name", "", "datastore-12", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewClient(u, user, passwd, true)
			if err != nil {
				t.Fatalf("failed %v", err)
			}
			defer func() { _ = c.Logout() }()

			err = c.DsSummary(tt.na, tt.moid, &LimitsStruct{}, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("DsMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
func TestClient_HostSummary(t *testing.T) {

	tests := []struct {
		name    string
		na      string
		moid    string
		wantErr bool
	}{
		{"moid", "", "host-540", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, user, passwd, true)
			if err != nil {
				t.Fatalf("failed %v", err)
			}
			defer func() { _ = c.Logout() }()

			err = c.HostSummary(tt.na, tt.moid, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("hostsummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
func TestClient_Metrics(t *testing.T) {
	tests := []struct {
		name     string
		prop     types.ManagedObjectReference
		metrics  []string
		interval int32
		wantErr  bool
	}{
		{"vm", types.ManagedObjectReference{Type: "VirtualMachine", Value: "vm-1087"}, vmSummaryDefault, 20, false},
		{"host", types.ManagedObjectReference{Type: "HostSystem", Value: "host-540"}, hsSummaryDefault, 20, false},
		{"ds", types.ManagedObjectReference{Type: "Datastore", Value: "datastore-12"}, dsSummaryDefault, 20, false},
		{"vds", types.ManagedObjectReference{Type: "VmwareDistributedVirtualSwitch", Value: "dvs-75"}, vdsSummaryDefault, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewClient(u, user, passwd, true)
			if err != nil {
				t.Fatalf("failed %v", err)
			}
			defer func() { _ = c.Logout() }()

			pr := newPrtgData("testing")
			err = c.Metrics(tt.prop, pr, tt.metrics, tt.interval)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			_ = pr.print(500*time.Microsecond, false)
		})
	}
}

func TestClient_VdsSummary(t *testing.T) {
	type args struct {
		searchName, searchMoid string
		txt                    bool
	}

	tests := []struct {
		name    string
		ur      *url.URL
		args    args
		wantErr bool
	}{
		{"", u, args{"DSwitch", "", false}, false},
		{"", u, args{"", "dvs-75", false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, user, passwd, true)
			if err != nil {
				t.Errorf("failed %v", err)
			}
			defer func() { _ = c.Logout() }()

			pr := newPrtgData("testing")

			err = c.VdsSummary(tt.args.searchName, tt.args.searchMoid, tt.args.txt)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			for i := range []int{0, 0, 0, 0, 0} {
				if i <= len(pr.items)-1 {
					printJSON(false, pr.items[i])
				}

			}
		})
	}
}

//func TestClient_hostMetrics(t *testing.T) {
//	u, err := url.Parse("https://192.168.0.201/sdk")
//	if err != nil {
//		t.Fatalf("failed to parse url")
//	}
//	tests := []struct {
//		name     string
//		na, moid string
//		u        *url.URL
//		wantErr  bool
//	}{
//		{"", "name", "", u, true},
//		{"", "*2", "", u, false},
//		{"", "", "Host-9", u, false},
//
//		//{"", property.Filter{"self": "*6"}, false},
//		//{"", property.Filter{"self": "VirtualMachine:vm-26"}, false},
//		//{"", property.Filter{"name": "*2"}, true},
//		//{"", nil, true},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c, err := NewClient(tt.u, "prtg@heynes.local", ".l3tm31n",timeout)
//			if err != nil {
//				t.Fatalf("failed %v", err)
//			}
//
//			gotM, err := c.hostSummary(tt.na, tt.moid, &LimitsStruct{}, true)
//			if (err != nil) != tt.wantErr {
//				t.Fatalf("hostSummary() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//
//			t.Logf(" %v", gotM)
//		})
//	}
//}

func TestClient_VmTracker(t *testing.T) {

	tests := []struct {
		name    string
		v, h    string
		wantErr bool
	}{
		{"", "vcenter", "192.168.0.1", false},
		{"", "mh-cache", "192.168.0.1", false},
		{"", "testServer", "192.168.0.1", false},
		{"", "testServer", "192.168.0.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{}
			if err := c.vmTracker(tt.v, tt.h); (err != nil) != tt.wantErr {
				t.Errorf("vmTracker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
