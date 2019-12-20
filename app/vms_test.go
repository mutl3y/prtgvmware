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
	"testing"
	"time"
)

var timeout = 2 * time.Second
var u, _ = url.Parse("https://192.168.0.201/sdk")

//func TestClient_vmSummary(t *testing.T) {
//	type args struct {
//		searchType, searchItem string
//		usr, pw                string
//		txt                    bool
//	}
//	u, err := url.Parse("https://192.168.0.201/sdk")
//	if err != nil {
//		t.Fatalf("failed to parse url")
//	}
//
//	tests := []struct {
//		name    string
//		ur      *url.URL
//		args    args
//		wantErr bool
//	}{
//		{"1", &url.URL{}, args{"name", "*1", "", "", false}, false},
//		{"2", &url.URL{}, args{"self", "*vm-30", "", "", false}, false},
//		{"3", &url.URL{}, args{"name", "me", "", "", false}, true},
//		//{"4", u, args{"name", "vcenter", "ps@heynes.local", ".l3tm31n", true}, false},
//		//{"5", u, args{"name", "vcenter", "ps@heynes.local", ".l3tm31n", true}, false},
//		{"6", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", false}, false},
//		{"5", u, args{"name", "vcenter", "prtg@heynes.local", ".l3tm31n", false}, false},
//		//{"6", u, args{"tags", "windows", "ps@heynes.local", ".l3tm31n", true}, false},
//	}
//	//	debug = true
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c, err := NewClient(tt.ur, tt.args.usr, tt.args.pw)
//			if err != nil {
//				t.Errorf("%+v", err)
//			}
//			f := property.Filter{tt.args.searchType: tt.args.searchItem}
//			lim := &LimitsStruct{}
//			err = c.VmSummary(f, lim, time.Hour, tt.args.txt)
//			if (err != nil) && !tt.wantErr {
//				t.Fatal(err)
//			}
//		})
//	}
//}

func TestClient_vmSummary(t *testing.T) {
	type args struct {
		searchName, searchMoid string
		usr, pw                string
		txt                    bool
	}

	tests := []struct {
		name    string
		ur      *url.URL
		args    args
		wantErr bool
	}{
		//{"1", &url.URL{}, args{"name", "*1", "", "", false}, false},
		//{"2", &url.URL{}, args{"self", "*vm-30", "", "", false}, false},
		//{"3", &url.URL{}, args{"name", "me", "", "", false}, true},
		////{"4", u, args{"name", "vcenter", "ps@heynes.local", ".l3tm31n", true}, false},
		////{"5", u, args{"name", "vcenter", "ps@heynes.local", ".l3tm31n", true}, false},
		//{"6", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", false}, false},
		{"6", u, args{"ad", "", "prtg@heynes.local", ".l3tm31n", false}, false},
		{"5", u, args{"vcenter", "", "prtg@heynes.local", ".l3tm31n", false}, false},

		//{"6", u, args{"tags", "windows", "ps@heynes.local", ".l3tm31n", true}, false},
	}
	//	debug = true
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.ur, tt.args.usr, tt.args.pw, true)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			lim := &LimitsStruct{}
			err = c.VmSummary(tt.args.searchName, tt.args.searchMoid, lim, time.Hour, tt.args.txt, []string{"cpu.ready.summation"})
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
				CreateTime:     time.Now().Truncate(time.Microsecond),
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
						CreateTime:        time.Now().Truncate(time.Microsecond),
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
						CreateTime:     time.Now().Truncate(time.Microsecond),
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
								CreateTime:        time.Now().Truncate(time.Microsecond),
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
				CreateTime:     time.Now().Truncate(time.Microsecond),
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
						CreateTime:        time.Now().Truncate(time.Microsecond),
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
						CreateTime:     time.Now().Truncate(time.Microsecond),
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
								CreateTime:        time.Now().Truncate(time.Microsecond),
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
			_, err := snapshotCount(time.Now(), tt.args.snp)
			if err != nil {
				b.Fatalf("failed %v", err)
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
		//{"1", &url.URL{}, args{"name", "1", "", "", false}, false},
		//{"2", &url.URL{}, args{"self", "vm-27", "", "", false}, false},
		//{"3", &url.URL{}, args{"name", "me", "", "", false}, true},
		//{"4", u, args{"name", "vcenter", "ps@heynes.local", ".l3tm31n", true}, false},
		{"5", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", []string{"windows", "PRTG"}, false}, false},
		{"6", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", []string{"windowsx"}, false}, true},
		//{"7", u, args{"tags", "windows", "ps@heynes.local", ".l3tm31n", true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewClient(tt.ur, tt.args.usr, tt.args.pw, true)
			if err != nil {
				t.Fatalf("%+v", err)
			}
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

//func TestClient_vmMstrics(t *testing.T) {
//	tests := []struct {
//		name    string
//		prop    property.Filter
//		wantErr bool
//	}{
//		{"", property.Filter{"name": "ad"}, false},
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
//			printJson(false, mets)
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
		//	{"", "", true},
		{"fail", "", "datastore-1", true},
		{"name", "raid5", "", false},
		//{"moid", "", "datastore-10", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", false)
			if err != nil {
				t.Fatalf("failed %v", err)
			}
			defer c.Logout()
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
		//	{"", "", true},
		//{"fail", "", "datastore-1", true},
		//{"name", "192.168.0.194", "", false},
		{"moid", "", "host-63", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "testsave", ".l3tm31n", true)

			//	c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n",false)
			if err != nil {
				t.Fatalf("failed %v", err)
			}
			err = c.HostSummary(tt.na, tt.moid, &LimitsStruct{}, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("hostsummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
func TestClient_Metrics(t *testing.T) {
	tests := []struct {
		name    string
		prop    types.ManagedObjectReference
		wantErr bool
	}{
		//{"", types.ManagedObjectReference{Type:  "VirtualMachine",Value: "vm-16"}, false},
		//{"", types.ManagedObjectReference{Type: "HostSystem", Value: "host-12"}, false},
		//{"ds", types.ManagedObjectReference{Type: "Datastore", Value: "datastore-10"}, false},
		{"vds", types.ManagedObjectReference{Type: "VmwareDistributedVirtualSwitch", Value: "dvs-19"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", false)
			if err != nil {
				t.Fatalf("failed %v", err)
			}
			pr := NewPrtgData("testing")
			err = c.MetricS(tt.prop, pr, append(vmSummaryDefault, vdsSummaryDefault...), 20)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func TestClient_VdsSummary(t *testing.T) {
	type args struct {
		searchName, searchMoid string
		usr, pw                string
		txt                    bool
	}

	tests := []struct {
		name    string
		ur      *url.URL
		args    args
		wantErr bool
	}{
		//{"1", &url.URL{}, args{"name", "*1", "", "", false}, false},
		//{"2", &url.URL{}, args{"self", "*vm-30", "", "", false}, false},
		//{"3", &url.URL{}, args{"name", "me", "", "", false}, true},
		////{"4", u, args{"name", "vcenter", "ps@heynes.local", ".l3tm31n", true}, false},
		////{"5", u, args{"name", "vcenter", "ps@heynes.local", ".l3tm31n", true}, false},
		//{"6", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", false}, false},
		{"6", u, args{"DSwitch", "", "prtg@heynes.local", ".l3tm31n", false}, false},
		{"7", u, args{"", "dvs-19", "prtg@heynes.local", ".l3tm31n", false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", false)
			if err != nil {
				t.Errorf("failed %v", err)
			}
			defer c.Logout()
			pr := NewPrtgData("testing")

			err = c.VdsSummary(tt.args.searchName, tt.args.searchMoid, &LimitsStruct{}, tt.args.txt)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			for i := range []int{0, 0, 0, 0, 0} {
				if i <= len(pr.items)-1 {
					printJson(false, pr.items[i])
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
//		{"", "", "host-9", u, false},
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
