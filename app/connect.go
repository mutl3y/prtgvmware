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
	"flag"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/simulator/vpx"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net/url"
	"os"
	"strings"
	"time"
)

// getEnvString returns string from environment variable.
func getEnvString(v string, def string) string {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	return r
}

// getEnvBool returns boolean from environment variable.
func getEnvBool(v string, def bool) bool {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	switch strings.ToLower(r[0:1]) {
	case "t", "y", "1":
		return true
	}

	return false
}

var (
	envURL      = "GOVMOMI_URL"
	envUserName = "GOVMOMI_USERNAME"
	envPassword = "GOVMOMI_PASSWORD"
	envInsecure = "GOVMOMI_INSECURE"
)

var urlDescription = fmt.Sprintf("ESX or vCenter URL [%s]", envURL)
var urlFlag = flag.String("url", getEnvString(envURL, ""), urlDescription)

var insecureDescription = fmt.Sprintf("Don't verify the server's certificate chain [%s]", envInsecure)
var insecureFlag = flag.Bool("insecure", getEnvBool(envInsecure, false), insecureDescription)

func processOverride(u *url.URL, envU, envP string) {

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

type Client struct {
	c   *vim25.Client
	r   *rest.Client
	m   *view.Manager
	ctx context.Context
}

type conView struct {
	*view.ContainerView
}

func NewClient(u *url.URL, user, pw string) (c Client, err error) {
	ctx := context.Background()
	if u.Host == "" {
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
		err := si.Create()
		if err != nil {
			return c, err
		}

		s := si.Service.NewServer()

		u = s.URL
		processOverride(u, "", "")
		soapClient := soap.NewClient(u, true)
		c.c, err = vim25.NewClient(ctx, soapClient)
		if err != nil {
			return c, fmt.Errorf("vim client sim %v", err)
		}
		//c.r = rest.NewClient(c.c)
		// run simulator for 10 seconds
		time.AfterFunc(10*time.Second, s.Close)
	} else {

		soapClient := soap.NewClient(u, true)
		c.c, err = vim25.NewClient(ctx, soapClient)
		if err != nil {
			return c, fmt.Errorf("unable to connect to %v ", u)
		}
		c.r = rest.NewClient(c.c)

	}
	processOverride(u, user, pw)
	err = sessionLogin(c.c, u)
	if err != nil {
		return Client{}, err
	}
	ui := url.UserPassword(user, pw)
	if c.r != nil {
		err = c.r.Login(ctx, ui)
		if err != nil {
			return c, fmt.Errorf("rest login %v", err)
		}

		ses, err := c.r.Session(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if ses == nil {
			log.Fatal("expected non-nil session")
		}
	}
	err = sessionCheck(c.c)
	if err != nil {
		return Client{}, err
	}
	c.m = view.NewManager(c.c)
	c.ctx = ctx
	return c, err
}

func sessionLogin(c *vim25.Client, ur *url.URL) error {
	req := types.Login{
		This: *c.ServiceContent.SessionManager,
	}

	req.UserName = ur.User.Username()
	if pw, ok := ur.User.Password(); ok {
		req.Password = pw
	}

	_, err := methods.Login(context.Background(), c, &req)
	if err != nil {
		return fmt.Errorf("login %v", err)
	}
	return nil
}

func sessionCheck(c *vim25.Client) error {
	var mgr mo.SessionManager

	err := mo.RetrieveProperties(context.Background(), c, c.ServiceContent.PropertyCollector, *c.ServiceContent.SessionManager, &mgr)
	if err != nil {
		return fmt.Errorf("session check %v", err)
	}
	return nil
}

func (c *Client) Logout() error {
	req := types.Logout{
		This: *c.c.ServiceContent.SessionManager,
	}
	c.c.CloseIdleConnections()
	_, err := methods.Logout(c.ctx, c.c, &req)
	if err != nil {
		return fmt.Errorf("logout %v", err)
	}
	if c.r != nil {
		_ = c.r.Logout(c.ctx)
		_ = c.r.CloseIdleConnections
	}

	return nil
}

//func Run(f func(context.Context, *vim25.Client) error) error {
//	flag.Parse()
//	var err error
//	if *urlFlag == "" {
//		s := sim{}
//		m := simulator.Model{
//			ServiceContent: vpx.ServiceContent,
//			RootFolder:     vpx.RootFolder,
//			Autostart:      true,
//			Datacenter:     1,
//			Portgroup:      1,
//			Host:           0,
//			Cluster:        1,
//			ClusterHost:    1,
//			Datastore:      1,
//			Machine:        2,
//			DelayConfig: simulator.DelayConfig{
//				Delay:       0,
//				DelayJitter: 0,
//				MethodDelay: nil,
//			},
//		}
//		s.Model = m
//		err = s.Run2(f)
//	} else {
//		ctx := context.Background()
//		c, err := NewClient(ctx)
//		if err == nil {
//			err = f(ctx, c.Client)
//		}
//	}
//	return err
//}

type sim struct {
	simulator.Model
	ctx    context.Context
	client *vim25.Client
}

//func NewSim()sim{
//	s := sim{}
//	s.ctx = context.Background()
//
//	s.client = NewClient()
//	return s
//}

func (m *sim) Run2(f func(context.Context, *vim25.Client) error) error {
	ctx := context.Background()

	defer m.Remove()
	err := m.Create()
	if err != nil {
		return err
	}

	s := m.Service.NewServer()
	defer s.Close()

	c, err := govmomi.NewClient(ctx, s.URL, true)
	if err != nil {
		return err
	}

	defer c.Logout(ctx)

	return f(ctx, c.Client)
}
