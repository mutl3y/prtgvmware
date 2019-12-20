// +build

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
	"fmt"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/simulator/vpx"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"net/url"
	"time"
)

var (
	envURL      = "GOVMOMI_URL"
	envUserName = "GOVMOMI_USERNAME"
	envPassword = "GOVMOMI_PASSWORD"
	envInsecure = "GOVMOMI_INSECURE"
)

type sim struct {
	simulator.Model
	ctx    context.Context
	client *vim25.Client
}

func _processOverride(u *url.URL, envU, envP string) {

	// Override username if provided
	if envU != "" {
		var password string
		var ok bool

		if u.User != nil {
			password, ok = u.User.Password()
		}

		if ok {
			u.User = url.UserPassword(envU, password)
		} else {
			u.User = url.User(envU)
		}
	}

	// Override password if provided
	if envP != "" {
		var username string

		if u.User != nil {
			username = u.User.Username()
		}

		u.User = url.UserPassword(username, envP)
	} else {
		u.User = url.UserPassword(envUserName, envPassword)
	}
}

// uncomment this to inject a simulator, removed to save 6MB in executable size
func runSim(ctx context.Context) (c Client, err error) {
	fmt.Println("Running in simulation mode, specify a url if you want this too work for real")
	si := sim{}
	m := &simulator.Model{
		ServiceContent: vpx.ServiceContent,
		RootFolder:     vpx.RootFolder,
		Autostart:      true,
		Datacenter:     1,
		Portgroup:      1,
		Host:           0,
		Cluster:        1,
		ClusterHost:    1,
		Datastore:      2,
		Machine:        2,
		DelayConfig: simulator.DelayConfig{
			Delay:       0,
			DelayJitter: 0,
			MethodDelay: nil,
		},
	}
	si.Model = *m
	defer si.Remove()
	err = si.Create()
	if err != nil {
		return
	}

	s := si.Service.NewServer()

	u := s.URL
	_processOverride(u, "", "")
	soapClient := soap.NewClient(u, true)
	c.c, err = vim25.NewClient(ctx, soapClient)
	if err != nil {
		return c, fmt.Errorf("vim client sim %v", err)
	}
	//c.r = rest.NewClient(c.c)
	// run simulator for 10 seconds
	time.AfterFunc(10*time.Second, s.Close)
	return
}
