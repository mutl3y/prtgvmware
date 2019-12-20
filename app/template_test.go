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
	"testing"
	"time"
)

func TestGenTemplate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := GenTemplate([]string{"prtg"}, time.Second, "deleteme")
			if err != nil {
				t.Fatalf("failed %v", err)
			}
		})
	}
}
func TestClient_DynTemplate(t *testing.T) {
	tests := []struct {
		name string
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(u, "prtg@heynes.local", ".l3tm31n", true)
			if err != nil {
				t.Errorf("failed %v", err)
			}

			err = c.DynTemplate([]string{"PRTG"}, time.Second, "deleteme")
			if err != nil {
				t.Fatalf("failed %v", err)
			}
		})
	}
}
