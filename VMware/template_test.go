package VMware

import (
	"testing"
)

func TestGenTemplate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := GenTemplate([]string{"windows", "PRTG"})
			if err != nil {
				t.Fatalf("failed %v", err)
			}
		})
	}
}
