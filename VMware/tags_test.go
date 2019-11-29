package VMware

import (
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"testing"
)

func TestClient_tagList(t *testing.T) {
	u, err := url.Parse("https://192.168.59.4/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}
	tests := []struct {
		name       string
		tagIds     []string
		wantRtnMap map[string]obJdata
		wantErr    bool
	}{
		{"1", []string{"windows"}, map[string]obJdata{"vm-15": {[]string{"windows"}, "vm-13", "VirtualMachine"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n")
			if err != nil {
				t.Fatal("cant get client")
			}
			gotRtnMap := NewTagMap()
			err = c.list(tt.tagIds, gotRtnMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("list() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func Test_tagMap_add(t *testing.T) {
	tm := NewTagMap()
	type args struct {
		vm  string
		tag string
	}
	tests := []struct {
		name    string
		args    args
		count   int
		wantErr bool
	}{
		{"", args{vm: "test", tag: "first"}, 1, false},
		{"", args{vm: "test2", tag: "second"}, 1, false},
		{"", args{vm: "test", tag: "first"}, 2, false},
		{"", args{vm: "test", tag: "second"}, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {

			oj := types.ManagedObjectReference{
				Type:  "VirtualMachine",
				Value: "",
			}
			oj.Value = tt.args.vm
			tm.add(oj, tt.args.tag)
			dat := tm.Data[tt.args.vm].Tags
			if (len(dat) != tt.count) && !tt.wantErr {
				t.Fatalf("wanted count of %v got %v \n%+v", tt.count, len(dat), dat)
			}
		})
	}
}

func Test_tagMap_check(t *testing.T) {
	u, err := url.Parse("https://192.168.59.4/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}
	tests := []struct {
		name, objId, tag string
		found, wantErr   bool
	}{
		{"1", "vm-15", "windows", true, false},
		{"2", "vm-13", "PRTG", true, false},
		{"3", "vm-15", "PRTG", true, false},
		{"4", "vm-15", "ARMAGENDON", false, true},
		{"5", "vm-19", "PRTG", false, true},
	}
	c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n")
	if err != nil {
		t.Fatal("cant get client")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttagMap := NewTagMap()
			err := c.list([]string{tt.tag}, ttagMap)
			if (err != nil) && !tt.wantErr {
				t.Fatalf("taglist error %v", err)
			}
			if (len(ttagMap.Data) == 0) && !tt.wantErr {
				t.Fatal("no Data returned")
			}

		})
	}
}

func TestClient_GetVmsOnTags(t *testing.T) {
	u, err := url.Parse("https://192.168.59.4/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}

	gotRtnMap := NewTagMap()
	tests := []struct {
		name    string
		tag     string
		wantErr bool
	}{
		{"", "PRTG", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n")
			if err != nil {
				t.Fatal("cant get client")
			}
			if err := c.GetObjIds(tt.tag, gotRtnMap); (err != nil) != tt.wantErr {
				t.Errorf("GetVmsOnTags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	t.Logf("%+v", gotRtnMap)
}
