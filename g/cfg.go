package g

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type SSHHost struct {
	Host     string
	Port     int
	Username string
	Password string
	Cmds     string
	Result   SSHResult
}

type HostJson struct {
	SSHHosts []SSHHost
}

type SSHResult struct {
	Host    string
	Success bool
	Result  string
}

func SplitString(str string) (strList []string) {
	if strings.Contains(str, ",") {
		strList = strings.Split(str, ",")
	} else if strings.Contains(str, ";") {
		strList = strings.Split(str, ";")
	} else {
		strList = strings.Split(str, "\n")
	}
	return
}

func GetFileAll(filePath string) ([]byte, error) {
	result, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("read file", filePath, err)
		return result, err
	}
	return result, err
}

func GetFile(filePath string) ([]string, error) {
	result := []string{}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("read file", filePath, err)
		return result, err
	}
	s := string(b)
	for _, lineStr := range strings.Split(s, "\n") {
		lineStr = strings.TrimSpace(lineStr)
		if lineStr == "" {
			continue
		}
		result = append(result, lineStr)
	}
	return result, nil
}

func GetJsonFile(filePath string) ([]SSHHost, error) {
	result := []SSHHost{}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("read file", filePath, err)
	}
	var m HostJson
	err = json.Unmarshal(b, &m)
	if err != nil {
		log.Println("read file", filePath, err)
		return result, err
	}
	result = m.SSHHosts
	return result, nil
}

func WriteIntoText(sshResult SSHResult, locate string) error {
	outputFile, outputError := os.OpenFile(locate+sshResult.Host+"txt", os.O_WRONLY|os.O_CREATE, 0666)
	if outputError != nil {
		return outputError
	}
	defer outputFile.Close()

	outputWriter := bufio.NewWriter(outputFile)
	outputString := sshResult.Result
	outputWriter.WriteString(outputString)
	outputWriter.Flush()
	return nil
}

func GetIPList(ipString string) ([]string, error) {
	res := SplitString(ipString)
	var allIP []string
	if len(res) > 0 {
		for _, sip := range res {
			aip := ParseIP(sip)
			for _, ip := range aip {
				allIP = append(allIP, ip)
			}
		}
	}
	return allIP, nil
}

func ParseIP(ip string) []string {
	var availableIPs []string
	ip = strings.TrimRight(ip, "/")
	if strings.Contains(ip, "/") == true {
		if strings.Contains(ip, "32") == true {
			aip := strings.Replace(ip, "/32", "", -1)
			availableIPs = append(availableIPs, aip)
		} else {
			availableIPs = GetAvailableIP(ip)
		}
	} else if strings.Contains(ip, "-") == true {
		ipRange := strings.SplitN(ip, "-", 2)
		availableIPs = GetAvailableIPRange(ipRange[0], ipRange[1])
	} else {
		availableIPs = append(availableIPs, ip)
	}
	return availableIPs
}

func GetAvailableIPRange(ipStart, ipEnd string) []string {
	var availableIPs []string
	firstIP := net.ParseIP(ipStart)
	endIP := net.ParseIP(ipEnd)
	if firstIP.To4() == nil || endIP.To4() == nil {
		return availableIPs
	}
	firstIPNum := ipToInt(firstIP.To4())
	endIPNum := ipToInt(endIP.To4())
	pos := int32(1)

	newNum := firstIPNum

	for newNum <= endIPNum {
		availableIPs = append(availableIPs, intToIP(newNum).String())
		newNum = newNum + pos
	}
	return availableIPs
}

func ipToInt(ip net.IP) int32 {
	return int32(binary.BigEndian.Uint32(ip.To4()))
}

func intToIP(n int32) net.IP {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return net.IP(b)
}

func GetAvailableIP(ipAndMask string) []string {
	var availableIPs []string

	ipAndMask = strings.TrimSpace(ipAndMask)
	ipAndMask = IPAddressToCIDR(ipAndMask)
	_, ipnet, _ := net.ParseCIDR(ipAndMask)
	firstIP, _ := networkRange(ipnet)
	ipNum := ipToInt(firstIP)
	size := networkSize(ipnet.Mask)
	pos := int32(1)
	max := size - 2

	var newNum int32
	for attempt := int32(0); attempt < max; attempt++ {
		newNum = ipNum + pos
		pos = pos%max + 1
		availableIPs = append(availableIPs, intToIP(newNum).String())
	}
	return availableIPs
}

func IPAddressToCIDR(ipAdress string) string {
	if strings.Contains(ipAdress, "/") == true {
		ipAndMask := strings.Split(ipAdress, "/")
		ip := ipAndMask[0]
		mask := ipAndMask[1]
		if strings.Contains(mask, ".") == true {
			mask = IPMaskStringToCIDR(mask)
		}
		return ip + "/" + mask
	} else {
		return ipAdress
	}
}

func IPMaskStringToCIDR(netmask string) string {
	netmaskList := strings.Split(netmask, ".")
	var mint []int
	for _, v := range netmaskList {
		strv, _ := strconv.Atoi(v)
		mint = append(mint, strv)
	}
	myIPMask := net.IPv4Mask(byte(mint[0]), byte(mint[1]), byte(mint[2]), byte(mint[3]))
	ones, _ := myIPMask.Size()
	return strconv.Itoa(ones)
}

func IPMaskCIDRToString(one string) string {
	oneInt, _ := strconv.Atoi(one)
	mIPmask := net.CIDRMask(oneInt, 32)
	var maskstring []string
	for _, v := range mIPmask {
		maskstring = append(maskstring, strconv.Itoa(int(v)))
	}
	return strings.Join(maskstring, ".")
}

func networkRange(network *net.IPNet) (net.IP, net.IP) {
	netIP := network.IP.To4()
	firstIP := netIP.Mask(network.Mask)
	lastIP := net.IPv4(0, 0, 0, 0).To4()
	for i := 0; i < len(lastIP); i++ {
		lastIP[i] = netIP[i] | ^network.Mask[i]
	}
	return firstIP, lastIP
}

func networkSize(mask net.IPMask) int32 {
	m := net.IPv4Mask(0, 0, 0, 0)
	for i := 0; i < net.IPv4len; i++ {
		m[i] = ^mask[i]
	}
	return int32(binary.BigEndian.Uint32(m)) + 1
}

func main() {}
