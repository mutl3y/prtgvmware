# prtgvmware
Tested with PRTG Version 19.3.51.2722

# Custom sensors for vmware stats for PRTG Network Monitor

* [Configuring PRTG Network Monitor](#configuring-prtg-network-monitor)
  * [Download](#download)
  * [Copy files](#copy-files)
  * [Adding device Metascan](#adding-device-using-metascan)
  * [Adding device Dynamic](#adding-device-using-dynamic-templates)
  * [Cached Credentials](#cached-credentials)
  * [Investigating issues](#investigating-issues)
  * [XML: The returned xml does not match the expected schema. (code: PE233)](#xml-the-returned-xml-does-not-match-the-expected-schema-code-pe233)

## Configuring PRTG Network Monitor

### Download
To install this prtg VMware sensor plugin you need to download the following
from the [releases](https://github.com/mutl3y/prtgvmware) section:
* plugin binary (choose correct binary according to your architecture) 
can be run remotely on a linux box, 
place on remote Host in standard vmware script location
on linux this is /var/prtg/scriptsxml/ and make the file executable via the user you intend to run it as 
remembering to enter, linux user creds in for remote Host

for windows place in the customsensosrs\exexml folder
**Make sure to download files from the latest release.**

## Actions
auto building of releases is was broken as some recent features of the govmomi vmware module are required that are not in the current release

I have included vendored packages that have been updated should you build this for yourself 

## Generate documentation 
```prtgvmware.exe GenDoc```

## Generate Template
This is for auto discovery purposes, Please see /docs for further info or run command with -h for further details of what you can use here

eg
```
prtgvmware.exe template--tags prtg --snapAge 7d
```

## Generate dynamic templates
This is for also auto discovery purposes, 

this gets round the way prtg metascan works and allows subsequent scans to add newly found items when tags change, 
does not delete anything so if you have a lot of churn on vm's you tag you will need to cleanup regularly
eg
```
prtgvmware.exe dynamicTemplates --tags prtg --snapAge 7d
```

### Copy files
* copy `prtgvmware.odt` to `C:\Program Files (x86)\PRTG Network Monitor\devicetemplates`
* copy `prtgvmware.exe` to `C:\Program Files (x86)\PRTG Network Monitor\Custom Sensors\EXEXML`

### Adding device using metascan
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

### Adding device using dynamic templates
* tag items you wish too monitor
* run dynamicTemplate with these tags using -j to check output
* if happy the content returned appears to resemble your infrastructure put this command into a batch file without the -j 
* add line to batchfile to copy the template to the devicetemplates folder of your PRTG server and test 
* Start PRTG Enterprise Console or PRTG Network Monitor (Web UI)
* Right-click your probe (in case of single server installation - Local Probe) and choose ```"Add Device"```
* Enter device name (e.g. ```vcenter```)
* In the "IPv4 Address/DNS Name of vcenter (without https://),
  e.g. `vcenter.local..net`
* Choose "Automatic sensor creation using specific device tplate in ```"Sensor Management"```
* Choose ``prtgvmware`` device template and deselect all others, this will be whatever you set in batchfile defaults to prtgvmware
* Change schedule to check ```hourly``` 
* Enter VMware read only creds to windows user and password, leave domain empty I.E. ```administrator@vsphere.local```
* Click ```"Continue"``` button 
* subsequent runs of the batchfile will create new devices

NOTE: PRTG will continue to track any deleted items so you will need to clean these up

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