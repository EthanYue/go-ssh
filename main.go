package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mySSH/funcs"
	"github.com/mySSH/g"
)

func main() {
	var hostList []string
	sshHosts := []g.SSHHost{}

	ips := "10.220.17.105"
	hostList, _ = g.GetIPList(ips)

	cmds := "dis int brief\n"

	username := "NENET"
	password := "netM@163"
	port := 22

	for _, host := range hostList {
		hostStruct := g.SSHHost{
			Host:     host,
			Username: username,
			Password: password,
			Port:     port,
			Cmds:     cmds,
		}
		sshHosts = append(sshHosts, hostStruct)
	}

	chLimit := make(chan bool, 10)
	chs := make([]chan g.SSHResult, len(sshHosts))
	startTime := time.Now()
	log.Println("Multissh start")
	limitFunc := func(chLimit chan bool, ch chan g.SSHResult, host g.SSHHost) {
		funcs.Dossh(host.Username, host.Password, host.Host, host.Port, host.Cmds, ch)
		<-chLimit
	}
	for i, host := range sshHosts {
		chs[i] = make(chan g.SSHResult, 1)
		chLimit <- true
		go limitFunc(chLimit, chs[i], host)
	}

	sshResults := []g.SSHResult{}
	for _, ch := range chs {
		res := <-ch
		if res.Result != "" {
			sshResults = append(sshResults, res)
		}
	}

	endTime := time.Now()
	log.Println("Multissh finished, Process time", endTime.Sub(startTime), "Number of active ip is", len(sshHosts))

	// // output file
	// for _, sshResult := range sshResults {
	// 	err = g.WriteIntoText(sshResult, "output.txt")
	// 	if err != nil {
	// 		log.Println("write into txt error:", err)
	// 		return
	// 	}
	// 	return
	// }

	// // jsonify print
	// jsonResult, err := json.Marshal(sshResults)
	// if err != nil {
	// 	log.Println("json marshal error:", err)
	// }
	// fmt.Println(string(jsonResult))
	// return

	// normal print
	for _, sshResult := range sshResults {
		fmt.Println("host:", sshResult.Host)
		fmt.Println("============= Result =============")
		fmt.Println(sshResult.Result)
	}
}
