package main

import (
"flag"
"fmt"
"strconv"
"path/filepath"
"io/ioutil"
"strings"
"os"
"time"
"encoding/gob"
)

//stuct to hold stats located in /sys/block/<device>/stat
type diskStat struct {
	Read_ios, Read_merges, Read_sectors, Read_ticks, Write_ios, Write_merges, Write_sectors, Write_ticks, Inflight_ios, Io_ticks, Time_in_queue, Timestamp uint64
	Disk string
}

//stuct to hold our latentcy stats
type diskLatencies struct {
	Read_await, Write_await, Await float64
}

//map to hold our disks
var diskMap = map[string]diskStat{}

//file to hold our disk stats for computing iowait
const gobFile = "/tmp/zabbix_iostat.gob"

//read disk perf stats from /sys/block/<device>/stat
//note: disk partition stats are located in /sys/block/<device>/<partition>/stat
func readDiskStats(disk string) diskStat {

	statsPath	:=	"/sys/block/"+disk+"/stat"
	stats,err	:=	ioutil.ReadFile(statsPath)

	//obtain timestamp of stats reading
	timestamp := uint64(time.Now().Unix())

	if err != nil {
		panic(err)
	}
	return diskStats2Stuct( string(stats), &timestamp, disk )
} 

//convert our disk stats (in string form) to type diskstat and return
func diskStats2Stuct( diskStats string, timestamp *uint64, diskname string ) diskStat {
	
	var disk diskStat
	statArr := strings.Fields(string(diskStats))
	if len(statArr) == 11{
		disk.Disk 				= diskname 
		disk.Timestamp 			= *timestamp
		disk.Read_ios,_			= strconv.ParseUint(statArr[0],10,64)
		disk.Read_merges,_		= strconv.ParseUint(statArr[1],10,64) 
		disk.Read_sectors,_		= strconv.ParseUint(statArr[2],10,64)
		disk.Read_ticks,_		= strconv.ParseUint(statArr[3],10,64)
		disk.Write_ios,_		= strconv.ParseUint(statArr[4],10,64)
		disk.Write_merges,_		= strconv.ParseUint(statArr[5],10,64)
		disk.Write_sectors,_	= strconv.ParseUint(statArr[6],10,64)
		disk.Write_ticks,_		= strconv.ParseUint(statArr[7],10,64)
		disk.Inflight_ios,_		= strconv.ParseUint(statArr[8],10,64)
		disk.Io_ticks,_			= strconv.ParseUint(statArr[9],10,64)
		disk.Time_in_queue,_	= strconv.ParseUint(statArr[10],10,64)
	}
	return disk
}

//Write our stats to gob in /tmp for calculating latencies
func writestats( stats diskStat ){

	//try and open gob containing stats
	file, err := os.Open(gobFile)
	
	//if we've opened file remove old file and create new 
	//also load our map into var diskMap
	if err != nil { 
		//fmt.Println("waffles")
		file, err = os.Create( gobFile )
	}else {
		diskMap = readstats( stats.Disk )
		err = os.Remove( gobFile )
		//fmt.Println("pancakes")
		file, err = os.Create( gobFile )
	}
	
	//update stats on map with current
	diskMap[stats.Disk] = stats 
	//fmt.Println(diskMap)

	//write map to gob
	if err == nil {
		//fmt.Println("crepes")
		encoder := gob.NewEncoder( file )
		encoder.Encode( diskMap )
	}

	file.Close()
}

//read gob for calculating latencies
func readstats( disk string ) map[string]diskStat {
	
	var decodedMap map[string]diskStat
	
	file, err := os.Open(gobFile)

	if err == nil {
		decoder := gob.NewDecoder(file)
		decoder.Decode( &decodedMap )
		
		//fmt.Printf("%d\n", decodedMap )
		file.Close()
		
		return decodedMap
	}
	return decodedMap
}

func getlatency( disk string ) diskLatencies {

	//var for calculated latencies
	var	awaits diskLatencies
	
	//get prior stats sample:
	statsSample1 := readstats(disk)[disk]

	//get current stats sample:
	statsSample2 := readDiskStats(disk)

	//take two samples to allow us to calculate deltas
	//fmt.Printf("Get Sample: %s\n",disk )
	
	//write current stats to gob
	writestats( statsSample2 )

	//find avg latency for reads on disk (ms)
	if float64(statsSample2.Read_ios-statsSample1.Read_ios) > 0.0 {
		awaits.Read_await	= float64(statsSample2.Read_ticks-statsSample1.Read_ticks)/float64(statsSample2.Read_ios-statsSample1.Read_ios)
	}else{
		awaits.Read_await 	= 0.0
	}
	
	//find avg latency for writes on disk (ms)
	if float64(statsSample2.Write_ticks-statsSample1.Write_ticks)/float64(statsSample2.Write_ios-statsSample1.Write_ios) > 0.0{
		awaits.Write_await	= float64(statsSample2.Write_ticks-statsSample1.Write_ticks)/float64(statsSample2.Write_ios-statsSample1.Write_ios)
	}else{
		awaits.Write_await  = 0.0
	}
	
	//find avg latency for writes and reads on disk (ms)
	if float64((statsSample2.Read_ios+statsSample2.Write_ios)-(statsSample1.Read_ios+statsSample1.Write_ios)) > 0.0{
		awaits.Await 		= float64((statsSample2.Read_ticks+statsSample2.Write_ticks)-(statsSample1.Read_ticks+statsSample1.Write_ticks))/float64((statsSample2.Read_ios+statsSample2.Write_ios)-(statsSample1.Read_ios+statsSample1.Write_ios))
	}else{
		awaits.Await 		= 0.0
	}
	return awaits
}

//determine and return avg read/write latencies per disk
func getcurrentlatency( disk string ) diskLatencies {
	var	statsSample1	diskStat
	var	statsSample2	diskStat
	var	awaits			diskLatencies

	//take two samples to allow us to calculate deltas
	//fmt.Printf("Get Sample1: %s\n",disk )
	statsSample1 = readDiskStats(disk)
	//time.sleep - not workable, needed to import/use C function instead
	//fmt.Printf("Sample1: %s %d\n",disk,statsSample1 )
	
	time.Sleep( 1 * time.Second )	
	
	//fmt.Printf("Get Sample1: %s\n",disk)
	statsSample2 = readDiskStats(disk)
	//fmt.Printf("Sample2: %s;%d\n",disk,statsSample2 )

	//find avg latency for reads on disk (ms)
	if float64(statsSample2.Read_ios-statsSample1.Read_ios) > 0.0 {
		awaits.Read_await	= float64(statsSample2.Read_ticks-statsSample1.Read_ticks)/float64(statsSample2.Read_ios-statsSample1.Read_ios)
	}else{
		awaits.Read_await 	= 0.0
	}
	
	//find avg latency for writes on disk (ms)
	if float64(statsSample2.Write_ticks-statsSample1.Write_ticks)/float64(statsSample2.Write_ios-statsSample1.Write_ios) > 0.0{
		awaits.Write_await	= float64(statsSample2.Write_ticks-statsSample1.Write_ticks)/float64(statsSample2.Write_ios-statsSample1.Write_ios)
	}else{
		awaits.Write_await  = 0.0
	}
	
	//find avg latency for writes and reads on disk (ms)
	if float64((statsSample2.Read_ios+statsSample2.Write_ios)-(statsSample1.Read_ios+statsSample1.Write_ios)) > 0.0{
		awaits.Await 		= float64((statsSample2.Read_ticks+statsSample2.Write_ticks)-(statsSample1.Read_ticks+statsSample1.Write_ticks))/float64((statsSample2.Read_ios+statsSample2.Write_ios)-(statsSample1.Read_ios+statsSample1.Write_ios))
	}else{
		awaits.Await 		= 0.0
	}
	return awaits
}

//function for lld display all disks/partitions for monitoring
func diskdiscovery() (string, error) {

	//Build a json string to send out for zabbix lld
	disco := "{\"data\":["

	foo,_	:= filepath.Glob("/dev/[svh]d[a-z]*")
	for i,dev := range foo {

		//ssdcheck( strings.TrimPrefix(dev,"/dev/") )
		
		disco += fmt.Sprintf("{\"{#DISKNAME}\":\"%s\",\"{#DISKTYPE}\":\"%s\"}", strings.TrimPrefix(dev,"/dev/"), ssdcheck( strings.TrimPrefix(dev,"/dev/") ) )
		if i < len(foo)-1 {
			disco += fmt.Sprintf(",")
		}
	}
	disco += "]}"
	return disco, nil
}

//check if our disk in question is an ssd
func ssdcheck( disk string ) ( string ){

	if _, err := os.Stat("/sys/block/"+disk+"/queue/rotational"); err == nil {
		rotational,err :=	ioutil.ReadFile("/sys/block/"+disk+"/queue/rotational")

		if err != nil {
			panic(err)
		} else {
			if strings.TrimSpace(string(rotational)) == "1" {
				return "hdd"
			}else if strings.TrimSpace(string(rotational))  == "0" {
				return "ssd"
			}
		}
	}
	return "hdd"
}

func main() {
	// ./disk_stat_bin --disk=sda --metric=wiops - example of usage
	diskPtr := flag.String("disk", "", "disk to get\n\tex.) disk_stat_bin --disk=sda --metric=wiops - example of usage\n")
	operationPtr := flag.String("metric","","disk metric to retrieve -(wiops, riops, rawait, wawait, await, cur_rawait, cur_wawait, cur_await)\n\tex.) disk_stat_bin --disk=sda --metric=wiops - example of usage\n")
	discoveryPtr := flag.Bool("discovery",false, "bool to determine if we're going to spit out discovery\n\tex.) disk_stat_bin --discovery - example of usage\n")
	flag.Parse()

	if *discoveryPtr == true{
		disco,_ := diskdiscovery()
		fmt.Print(disco)
	} else if *operationPtr != "" && *diskPtr !="" {
		switch{
		case *operationPtr == "wiops":
			fmt.Print( readDiskStats( *diskPtr ).Write_ios )
		case *operationPtr == "riops":
			fmt.Print( readDiskStats( *diskPtr ).Read_ios )
		case *operationPtr == "rawait":
			fmt.Print( getlatency( *diskPtr ).Read_await )
		case *operationPtr == "wawait":
			fmt.Print( getlatency( *diskPtr ).Write_await )
		case *operationPtr == "await":
			fmt.Print( getlatency( *diskPtr ).Await )
		case *operationPtr == "cur_rawait":
			fmt.Print( getcurrentlatency( *diskPtr ).Read_await )
		case *operationPtr == "cur_wawait":
			fmt.Print( getcurrentlatency( *diskPtr ).Write_await )
		case *operationPtr == "cur_await":
			fmt.Print( getcurrentlatency( *diskPtr ).Await )
		}
	}
}