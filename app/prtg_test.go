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
	ps "github.com/PRTG/go-prtg-sensor-api"
	"sync"
	"testing"
	"time"
)

func TestPrtgData2_Print2(t *testing.T) {
	i := ps.SensorChannel{
		Channel:         "a test channel",
		Value:           "66",
		ValueMode:       "",
		Unit:            "",
		CustomUnit:      "",
		ValueLookup:     "",
		VolumeSize:      "",
		SpeedSize:       "",
		SpeedTime:       "",
		Float:           "1",
		DecimalMode:     "",
		ShowChart:       "1",
		ShowTable:       "1",
		LimitMinWarning: "",
		LimitMaxWarning: "",
		LimitWarningMsg: "",
		LimitMinError:   "",
		LimitMaxError:   "",
		LimitErrorMsg:   "",
		LimitMode:       "",
		Warning:         "",
	}

	items := make([]ps.SensorChannel, 0, 10)
	items = append(items, i)

	type fields struct {
		mu    *sync.RWMutex
		name  string
		moid  string
		err   string
		txt   string
		items []ps.SensorChannel
	}
	type args struct {
		checkTime time.Duration
		txt       bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"1", fields{
			mu:    nil,
			name:  "test",
			moid:  "vm01",
			err:   "",
			txt:   "",
			items: items,
		}, args{
			checkTime: 344,
			txt:       false,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PrtgData{
				mu:    tt.fields.mu,
				name:  tt.fields.name,
				moid:  tt.fields.moid,
				err:   tt.fields.err,
				text:  tt.fields.txt,
				items: tt.fields.items,
			}
			if err := p.Print(tt.args.checkTime, tt.args.txt); (err != nil) != tt.wantErr {
				t.Errorf("Print() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
