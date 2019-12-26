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
	"net/url"
	"testing"
)

func TestClient_Save2Disk(t *testing.T) {
	u, err := url.Parse("https://192.168.0.201/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}
	tests := []struct {
		name    string
		fn      string
		wantErr bool
	}{
		{"", "test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", true)
			if err != nil {
				t.Fatalf("failed %v", err)
			}
			defer func() { _ = c.Logout() }()
			if err := c.save2Disk(tt.fn, ".l3tm31n"); (err != nil) != tt.wantErr {
				t.Errorf("save2Disk() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewClientFromDisk(t *testing.T) {
	u, err := url.Parse("https://192.168.0.201/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}
	tests := []struct {
		name    string
		fn      string
		wantErr bool
	}{
		{"", "test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := clientFromDisk(tt.fn, ".l3tm31n", u)
			if (err != nil) != tt.wantErr {
				t.Errorf("clientFromDisk() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
