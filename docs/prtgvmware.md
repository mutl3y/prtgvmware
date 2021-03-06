## prtgvmware

VMware sensors for prtg

### Synopsis

advanced sensors for VMware

this app exposes all the common stats for vm's, Hypervisor's, VDS & Datastore's

to use autodiscovery you need to generate template using tags for each set of objects you want to monitor


### Options

```
  -c, --cachedCreds        disable cached connection
  -h, --help               help for prtgvmware
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

* [prtgvmware dsSummary](prtgvmware_dsSummary.md)	 - summary for a single datastore
* [prtgvmware dynamicTemplates](prtgvmware_dynamicTemplates.md)	 - generate prtg template for autodiscovery
* [prtgvmware genDocs](prtgvmware_genDocs.md)	 - Create documentation for app
* [prtgvmware hsSummary](prtgvmware_hsSummary.md)	 - summary for a single host
* [prtgvmware metascan](prtgvmware_metascan.md)	 - returns prtg sensors for autodiscovery
* [prtgvmware snapshots](prtgvmware_snapshots.md)	 - snapshots for many vm's
* [prtgvmware summary](prtgvmware_summary.md)	 - vm summary for a single machine
* [prtgvmware template](prtgvmware_template.md)	 - generate prtg template for autodiscovery
* [prtgvmware vdsSummary](prtgvmware_vdsSummary.md)	 - vds summary for prtg

###### Auto generated by spf13/cobra on 4-Feb-2020
