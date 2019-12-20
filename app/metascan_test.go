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
	"net/url"
	"testing"
	"time"
)

func TestClient_Metascan(t *testing.T) {
	u, err := url.Parse("https://192.168.0.201/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}

	tests := []struct {
		name string
		tags []string

		wantErr bool
	}{
		{"", []string{"windows", "PRTG"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx, _ = context.WithTimeout(ctx, time.Second)
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", false)
			if err != nil {
				t.Fatal("cant get client")
			}

			gotRtnMap := NewTagMap()
			if err := c.Metascan(tt.tags, gotRtnMap, time.Minute); (err != nil) != tt.wantErr {
				t.Errorf("Metascan() %v, wantErr %v", err, tt.wantErr)
			}
			err = c.Logout()
			if err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestClient_getObjType(t *testing.T) {
	u, err := url.Parse("https://192.168.0.201/sdk")
	if err != nil {
		t.Fatalf("failed to parse url")
	}

	tests := []struct {
		name string
		moid string

		wantErr bool
	}{
		{"ad", "vm-16", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", true)
			if err != nil {
				t.Fatal("cant get client")
			}

			moi := newMoidNames(&c)
			na := moi.GetName(tt.moid)
			if na == "" {
				t.Fatalf("could not find %v", na)
			}
		})
	}
}
