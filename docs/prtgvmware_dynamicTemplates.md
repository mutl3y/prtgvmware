## prtgvmware dynamicTemplates

generate prtg template for autodiscovery

### Synopsis

use this to support autodiscovery using VMware tags

run this regually via cron or task scheduler and copy template to devicetemplates folder for use by autodiscovery


```
prtgvmware dynamicTemplates [flags]
```

### Options

```
  -h, --help              help for dynamicTemplates
  -f, --template string   filename to save template as, adds .odt, only needed if using multiple sensors (default "prtgvmware")
```

### Options inherited from parent commands

```
  -c, --cachedCreds        disable cached connection (default true)
  -j, --json               pretty print json version of vmware data
      --maxErr float       greater than equal this will trigger a error response (used with snapshots)
      --maxWarn float      greater than equal this will trigger a warning response (used with snapshots) (default 1)
      --msgError string    message to use if error value exceeded (used with snapshots)
      --msgWarn string     message to use if warning value exceeded (used with snapshots)
  -n, --name string        name of vm, supports *partofname*
  -i, --oid string         exact id of an object e.g. vm-12, vds-81, host-9, datastore-10 
  -p, --password string    vcenter password
  -a, --snapAge duration   ignore snapshots younger than (default 168h0m0s)
  -t, --tags strings       slice of tags to include
  -U, --url string         url for vcenter api
  -u, --username string    vcenter username
```

### SEE ALSO

* [prtgvmware](prtgvmware.md)	 - VMware sensors for prtg

###### Auto generated by spf13/cobra on 20-Dec-2019