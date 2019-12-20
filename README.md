# PRTG_vmware
Tested with PRTG Version 19.3.51.2722

# Custom sensors for vmware stats for PRTG Network Monitor

* [Configuring PRTG Network Monitor](#configuring-prtg-network-monitor)
  * [Download](#download)
  * [Copy files](#copy-files)
  * [Adding device](#adding-device)
  * [Cached Credentials](#Cached Credentials)
  * [Investigating issues](#investigating-issues)
  * [XML: The returned xml does not match the expected schema. (code: PE233)](#xml-the-returned-xml-does-not-match-the-expected-schema-code-pe233)

## Configuring PRTG Network Monitor

### Download
To install this prtg VMware sensor plugin you need to download the following
from the [releases](https://github.com/mutl3y/PRTG_VMware) section:
* plugin binary (choose correct binary according to your architecture) 
can be run remotely, place on remote host in standard vmware script location
on linux this is /var/prtg/scriptsxml/ and make the file executable via the user you intend to run it as 
remembering to enter, linux user creds in for remote host

**Make sure to download files from the latest release.**

## Actions
auto building of releases is currently broken as some recent features of the govmomi vmware module are required that are not in the current release
if you do build this for yourself ensure you update govmomi separately to latest master version before compiling or it will fail to build

## Generate documentation 
```prtgvmware.exe GenDoc```

## Generate Template
This is for auto discovery purposes, Please see /docs for further info or run command with -h for further details of what you can use here

eg
```
prtgvmware.exe template--tags prtg --snapAge 7d
```

## Generate dynamic templates
This is for auto discovery purposes also, 

this gets round the way prtg metascan works and allows subsequent scans to add newly found items
eg
```
prtgvmware.exe dynamicTemplates --tags prtg --snapAge 7d
```

### Copy files
* copy `prtgvmware.odt` to `C:\Program Files (x86)\PRTG Network Monitor\devicetemplates`
* copy `prtgvmware.exe` to `C:\Program Files (x86)\PRTG Network Monitor\Custom Sensors\EXEXML`

### Adding device
* Start PRTG Enterprise Console or PRTG Network Monitor (Web UI)
* Right-click your probe (in case of single server installation - Local Probe) and choose ```"Add Device"```
* Enter device name (e.g. ```vcenter```)
* In the "IPv4 Address/DNS Name of vcenter (without https://),
  e.g. `vcenter.local..net`
* Choose "Automatic sensor creation using specific device tplate in ```"Sensor Management"```
* Choose ``prtgvmware`` device tplate and deselect all others
* Change schedule to check ```hourly``` 
* Enter VMware read only creds to windows user and password, leave domain empty I.E. ```administrator@vsphere.local```
* Click ```"Continue"``` button 

## Cached Credentials
Users connection is cached to file by default, this is encrypted using the supplied password, 
you can disable this behaviour causing the client to reconnect each time

## Investigating issues

##### XML: The returned xml does not match the expected schema. (code: PE233)

To investigate this issue please:
* click the failing sensor
* choose Settings button
* in the "SENSOR SETTINGS" section, item "EXE Result" mark "Write EXE result to disk"
* let the sensor run (wait for the period of execution, "Scanning Interval" on the same screen)
* review the files from `%programdata%\Paessler\PRTG Network Monitor\Logs (Sensors)\` sensorid.*
* review and please log issue if you suspect code is at fault 