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
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"
)

var pathSep = string(os.PathSeparator)

// Client holds the connections we need to query the remote system
type Client struct {
	Cached bool
	c      *vim25.Client
	r      *rest.Client
	m      *view.Manager
	ctx    context.Context
}

// NewClient returns a logged in client
func NewClient(u *url.URL, user, pw string, cache bool) (c Client, err error) {

	// load from cache if enabled, will fall through to login code if there are any issues
	if cache {
		c, err := clientFromDisk(u.Host, pw, u)
		if err == nil {
			c.Cached = true
			return c, nil
		}
	}

	ctx := context.Background()
	if u.Host == "" {
		return Client{}, fmt.Errorf("you need to provide a url, I.E https://vcenter/sdk")
	}

	u.User = url.UserPassword(user, pw)
	soapClient := soap.NewClient(u, true)
	c.c, err = vim25.NewClient(ctx, soapClient)
	if err != nil {
		return c, fmt.Errorf("unable to connect to %v ", u)
	}
	c.r = rest.NewClient(c.c)

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
	if cache {
		err := c.save2Disk(u.Host, pw)
		if err != nil {
			return c, fmt.Errorf("failed to save cached creds to disk")
		}
		c.Cached = true

	}
	return c, err
}

func configDir() string {
	var dir string

	switch runtime.GOOS {
	case "windows":
		var systemSettingFolders = []string{os.Getenv("PROGRAMDATA")}
		dir = strings.Join([]string{systemSettingFolders[0], "Paessler", "prtgvmware"}, pathSep)
	default:
		dir = strings.Join([]string{"/etc", "Paessler", "prtgvmware"}, pathSep)
	}
	return dir
}

func clientFromDisk(fn, password string, u *url.URL) (c Client, err error) {
	apiCookie := fn + ".api"
	restCookie := fn + ".rest"

	ccfn := strings.Join([]string{configDir(), apiCookie}, pathSep)
	rsfn := strings.Join([]string{configDir(), restCookie}, pathSep)

	// check files exist
	_, err = os.Stat(ccfn)
	if err != nil {
		fmt.Println("12")
		return
	}
	_, err = os.Stat(rsfn)
	if err != nil {
		fmt.Println("13")
		return
	}

	ctx := context.Background()

	// api clinet
	c = Client{}
	soapClient := soap.NewClient(u, true)
	c.c, err = vim25.NewClient(ctx, soapClient)
	if err != nil {
		return c, fmt.Errorf("unable to connect to %v ", u)
	}

	byc, err := ioutil.ReadFile(ccfn)
	bycDecrypted, err := Decrypt(byc, password)
	if err != nil {
		return Client{}, fmt.Errorf("could not decrypt creds: %v", err)
	}
	err = c.c.UnmarshalJSON(bycDecrypted)
	if err != nil {
		return Client{}, fmt.Errorf("read api cookie error: %v", err)
	}
	if c.c.URL().Host != u.Host {
		c.Cached = false
		return Client{}, fmt.Errorf("url mismatch, logging back in")
	}

	if !c.c.Valid() {
		return
	}
	if !c.c.IsVC() {
		return
	}
	err = sessionCheck(c.c)
	if err != nil {
		return Client{}, fmt.Errorf("failed session check: %v", err)

	}

	// rest client
	c.r = rest.NewClient(c.c)
	byr, err := ioutil.ReadFile(rsfn)
	byrDecrypted, err := Decrypt(byr, password)
	if err != nil {
		return Client{}, fmt.Errorf("could not decrypt creds: %v", err)
	}
	err = c.r.UnmarshalJSON(byrDecrypted)
	if err != nil {
		return Client{}, fmt.Errorf("read rest cookie error: %v", err)
	}

	c.ctx = ctx
	c.m = view.NewManager(c.c)
	return c, nil
}

func (c *Client) save2Disk(fn, password string) (err error) {
	apiCookie := fn + ".api"
	restCookie := fn + ".rest"

	// make sure we have a directory
	dir := configDir()
	err = os.MkdirAll(dir, 0644)
	if err != nil {
		return
	}

	// if we have a valid rest client
	if c.r != nil {
		rsfn := strings.Join([]string{configDir(), restCookie}, pathSep)
		rs, err := c.r.MarshalJSON()
		if err != nil {
			return err
		}
		rsEncrypt, err := Encrypt(rs, password)
		if err != nil {
			return err

		}
		err = ioutil.WriteFile(rsfn, rsEncrypt, 0644)
		if err != nil {
			return err
		}

	}

	// if we have a valid api client
	if c.c != nil {
		ccfn := strings.Join([]string{configDir(), apiCookie}, pathSep)

		cc, err := c.c.MarshalJSON()
		if err != nil {
			return err
		}
		ccEncrypt, err := Encrypt(cc, password)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(ccfn, ccEncrypt, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// Logout of remote system
func (c *Client) Logout() error {
	if c.Cached {
		return nil
	}
	req := types.Logout{
		This: *c.c.ServiceContent.SessionManager,
	}
	c.c.CloseIdleConnections()
	_, err := methods.Logout(context.Background(), c.c, &req)
	if err != nil {
		return fmt.Errorf("logout %v", err)
	}
	if c.r != nil {
		_ = c.r.Logout(c.ctx)
		_ = c.r.CloseIdleConnections
	}

	return nil
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

//func NewSim()sim{
//	s := sim{}
//	s.ctx = context.Background()
//
//	s.client = NewClient()
//	return s
//}
//
//func (m *sim) Run2(f func(context.Context, *vim25.Client) error) error {
//	ctx := context.Background()
//
//	defer m.Remove()
//	err := m.Create()
//	if err != nil {
//		return err
//	}
//
//	s := m.Service.NewServer()
//	defer s.Close()
//
//	c, err := govmomi.NewClient(ctx, s.URL, true)
//	if err != nil {
//		return err
//	}
//
//	defer func(){_=c.Logout()}()(ctx)
//
//	return f(ctx, c.Client)
//}
