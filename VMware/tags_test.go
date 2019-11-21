package VMware

import (
	"net/url"
	"reflect"
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
		wantRtnMap map[string][]string
		wantErr    bool
	}{
		{"1", []string{"windows", "PRTG"}, map[string][]string{"vm-15": {"windows", "PRTG"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n")
			if err != nil {
				t.Fatal("cant get client")
			}
			gotRtnMap := newTagMap()
			err = c.tagList(tt.tagIds, gotRtnMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("tagList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRtnMap.data, tt.wantRtnMap) {
				t.Errorf("tagList() gotRtnMap = %v, want %v", gotRtnMap, tt.wantRtnMap)
			}
		})
	}
}

func Test_tagMap_add(t *testing.T) {
	tm := newTagMap()
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

			tm.add(tt.args.vm, tt.args.tag)
			dat := tm.data[tt.args.vm]
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
			ttagMap := newTagMap()
			err := c.tagList([]string{tt.tag}, ttagMap)
			if err != nil {
				t.Fatalf("taglist error %v", err)
			}
			if (len(ttagMap.data) == 0) && !tt.wantErr {
				t.Fatal("no data returned")
			}

		})
	}
}
