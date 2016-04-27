package gogrinder

import (
	"fmt"
	"os"
)


// JtlReporter
// Jtl is the format used by Jmeter to exchange data with other Java tools like Jenkins
// https://wiki.apache.org/jmeter/JtlFiles
type JtlReporter struct {
	logfile *os.File
}


// Write metrics to the jtl result file.
func (r *JtlReporter) Update(m Metric) {
	// sample result.jtl file (usually there is no header!)
	// timeStamp,elapsed,label,responseCode,responseMessage,threadName,dataType,success,bytes,grpThreads,allThreads,Latency
	// 1461685566118,599,Home page,200,OK,User threads 1-25,text,true,18193,180,180,258
	// note: this is not a full implementation of the Jtl format
	// just the generic part we need to transfer results to Jenkins
	// for a full implementation we need to move this into every Reporter
	success := "true"
	if m.GetError() != "" { success = "false"}
	fmt.Fprintf(r.logfile, "%d,%d,%s,,,,text,%s,,,,\n",
		(m.GetTimestamp()).UnixNano()/1000000,
		m.GetElapsed()/1000000, m.GetTeststep(), success)
}
