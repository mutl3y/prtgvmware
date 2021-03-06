## prtgvmware summary

vm summary for a single machine

### Synopsis

queries vm performance metrics and outputs in PRTG format

Can be further expanded by adding additional VmWare performance counters

counters included by default are 
"cpu.readiness.average", "cpu.usage.average",
"datastore.datastoreNormalReadLatency.latest", "datastore.datastoreNormalWriteLatency.latest",
"datastore.datastoreReadIops.latest", "datastore.datastoreWriteIops.latest",
"disk.read.average", "disk.write.average", "disk.usage.average",
"mem.active.average", "mem.consumed.average", "mem.usage.average",
"net.bytesRx.average", "net.bytesTx.average", "net.usage.average",


```
prtgvmware summary [flags]
```

### Options

```
  -h, --help                help for summary
      --vmMetrics strings   include additional vm metrics, I.E. cpu.ready.summation
```

### Options inherited from parent commands

```
  -c, --cachedCreds        disable cached connection
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

###### Auto generated by spf13/cobra on 4-Feb-2020
