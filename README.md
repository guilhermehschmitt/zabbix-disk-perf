# zabbix-disk-perf
Binary written in go for use in discovering/monitoring disk(s) in linux via Zabbix agent & userparams

### Discovery 
Currently Disks are discovered via passing --discovery arg, which presents a listing of disks (and partions) in JSON, including their type (SSD or HDD) partitions can be excluded by regex on discovery ex) ^[sv]d[a-z]*$ .

```
./disk_stat_bin --discovery
{"data":[{"{#DISKNAME}":"vda","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vda1","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vda2","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vda5","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vdc","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vdc1","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vdd","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vdd1","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vde","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vde1","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vdf","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vdg","{#DISKTYPE}":"hdd"},{"{#DISKNAME}":"vdg1","{#DISKTYPE}":"hdd"}]}
```

### Graph generated from Data
![Alt text](/images/graph.png?raw=true "Graph of iops/latency generated from data" )
