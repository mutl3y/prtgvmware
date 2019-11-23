package VMware

import (
	"github.com/PRTG/go-prtg-sensor-api"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"testing"
	"time"
)

func TestClient_vmSummary(t *testing.T) {
	type args struct {
		searchType, searchItem string
		usr, pw                string
		txt                    bool
	}
	u, err := url.Parse("https://192.168.59.4/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}

	tests := []struct {
		name    string
		ur      *url.URL
		args    args
		wantErr bool
	}{
		{"1", &url.URL{}, args{"name", "*1", "", "", false}, false},
		{"2", &url.URL{}, args{"self", "*vm-30", "", "", false}, false},
		{"3", &url.URL{}, args{"name", "me", "", "", false}, true},
		//{"4", u, args{"name", "vcenter", "prtg@heynes.local", ".l3tm31n", true}, false},
		//{"5", u, args{"name", "vcenter", "prtg@heynes.local", ".l3tm31n", true}, false},
		{"6", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", false}, false},
		{"5", u, args{"name", "vcenter", "prtg@heynes.local", ".l3tm31n", false}, false},
		//{"6", u, args{"tags", "windows", "prtg@heynes.local", ".l3tm31n", true}, false},
	}
	//	debug = true
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.ur, tt.args.usr, tt.args.pw)
			if err != nil {
				t.Errorf("%+v", err)
			}
			f := property.Filter{tt.args.searchType: tt.args.searchItem}
			lim := &LimitsStruct{}
			err = c.VmSummary(f, lim, time.Hour, tt.args.txt)
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
	u, err := url.Parse("https://192.168.59.4/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
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
		//{"4", u, args{"name", "vcenter", "prtg@heynes.local", ".l3tm31n", true}, false},
		{"5", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", []string{"windows", "PRTG"}, true}, false},
		{"6", u, args{"name", "ad", "prtg@heynes.local", ".l3tm31n", []string{"windowsx"}, false}, true},
		//{"7", u, args{"tags", "windows", "prtg@heynes.local", ".l3tm31n", true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.ur, tt.args.usr, tt.args.pw)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			f := property.Filter{tt.args.searchType: "*" + tt.args.searchItem}
			lim := &LimitsStruct{}

			err = c.SnapShotsOlderThan(f, tt.args.tag, lim, time.Second, true)
			if (err != nil) && !tt.wantErr {
				t.Errorf("failed %v", err)
			}

		})
	}
}

func TestPrtgData_JSON(t *testing.T) {
	type fields struct {
		name  string
		err   string
		items map[string]Prtgitem
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "no error int",
			fields: fields{
				err: "",
				items: map[string]Prtgitem{"int": {
					Value: 1,
					Unit:  string(prtg.Count),
				}, "float": {
					Value: 443.212,
					Unit:  string(prtg.MegaBit),
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PrtgData{
				err:   tt.fields.err,
				items: tt.fields.items,
			}
			if err := p.Print(0, true); (err != nil) != tt.wantErr {
				t.Errorf("XML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_MetricSeries(t *testing.T) {
	tests := []struct {
		name    string
		prop    property.Filter
		wantErr bool
	}{
		{"", property.Filter{"name": "ad"}, false},
		//{"", property.Filter{"self": "*"}, true},
		//{"", property.Filter{"self": "*"}, true},
		//{"", property.Filter{"name": "*2"}, true},
		//{"", nil, true},
	}
	u, err := url.Parse("https://192.168.59.4/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n")
			if err != nil {
				t.Errorf("failed %v", err)
			}

			h, res, err := c.vmMetricS(tt.prop)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			t.Logf("%v %v", h, res)

		})
	}
}

//func TestClient_dsMetrics(t *testing.T) {
//	tests := []struct {
//		name    string
//		filter  property.Filter
//		wantErr bool
//	}{
//		{"", property.Filter{"name": "*1"}, false},
//		//{"", property.Filter{"self": "*6"}, false},
//		//{"", property.Filter{"self": "VirtualMachine:vm-26"}, false},
//		//{"", property.Filter{"name": "*2"}, true},
//		//{"", nil, true},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c, err := NewClient(&url.URL{}, "", "")
//			if err != nil {
//				t.Errorf("failed %v", err)
//			}
//
//			gotDs, gotM, err := c.dsMetrics("*1")
//			if (err != nil) != tt.wantErr {
//				t.Errorf("dsMetrics() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//
//			t.Logf("%v %v", gotDs, gotM)
//		})
//	}
//}
func TestClient_hostMetrics(t *testing.T) {
	tests := []struct {
		name    string
		filter  property.Filter
		wantErr bool
	}{
		{"", property.Filter{"name": "*1"}, false},
		//{"", property.Filter{"self": "*6"}, false},
		//{"", property.Filter{"self": "VirtualMachine:vm-26"}, false},
		//{"", property.Filter{"name": "*2"}, true},
		//{"", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(&url.URL{}, "", "")
			if err != nil {
				t.Errorf("failed %v", err)
			}

			gotDs, gotM, err := c.hostMetrics("*1")
			if (err != nil) != tt.wantErr {
				t.Errorf("hostMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			t.Logf("%v %v", gotDs, gotM)
		})
	}
}
