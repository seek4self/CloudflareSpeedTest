package task

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"CloudflareSpeedTest/utils"
)

const (
	hosts         = "C:\\Windows\\System32\\drivers\\etc\\hosts"
	defaultDomain = "localhost"
)

var Domain = defaultDomain

func ReplaceHosts() {
	if Domain == defaultDomain {
		fmt.Println("Please set domain!")
		return
	}
	ok, ip := readHosts(Domain)
	if !ok {
		fmt.Println("Read old IP faild")
		return
	}
	newIP := testNewIP(ip)
	if newIP == ip {
		fmt.Println("The current IP is fast enough ")
		return
	}
	if len(newIP) == 0 {
		return
	}
	replaceHosts(Domain, newIP)
}

func readHosts(domain string) (ok bool, ip string) {
	file, err := os.Open(hosts)
	if err != nil {
		fmt.Println("open hosts err", err)
		return
	}
	defer file.Close()
	scan := bufio.NewScanner(file)
	for scan.Scan() {
		s := scan.Text()
		if strings.Contains(s, domain) {
			ip = strings.Split(s, " ")[0]
			ok = true
			return
		}
	}
	fmt.Printf("Not found ip of %s\n", domain)
	return
}

func speed2MB(s float64) float64 {
	return s / 1024 / 1024
}

func checkData(delay time.Duration, speed float64) bool {
	return (delay < utils.InputMaxDelay && delay > utils.InputMinDelay) && speed > MinSpeed
}

func testNewIP(ip string) string {
	fmt.Println("Test old IP ...")
	addr := &net.IPAddr{IP: net.ParseIP(ip)}
	recv, delay := checkConnection(addr)
	avgDelay := delay / time.Duration(recv)
	avgSpeed := speed2MB(downloadHandler(addr))
	fmt.Printf("Old IP '%s' delay: %s speed: %.2f MB/s\n", ip, avgDelay, avgSpeed)
	if checkData(avgDelay, avgSpeed) {
		return ip
	}
	fmt.Println("Test new ip ...")
	var newIP utils.CloudflareIPData
	for {
		pingData := NewPing().Run().FilterDelay()
		speedData := TestDownloadSpeed(pingData)
		// speedData.Print(IPv6)
		if len(speedData) == 0 {
			fmt.Println("No IP Found")
			continue
		}
		newIP = speedData[0]
		if checkData(newIP.Delay, newIP.DownloadSpeed) {
			break
		}
	}
	fmt.Printf("New IP '%s' delay: %s speed: %.2f MB/s\n", newIP.IP.String(), newIP.Delay, speed2MB(newIP.DownloadSpeed))
	return newIP.IP.String()
}

func replaceHosts(domain, ip string) {
	f, err := os.OpenFile(hosts, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println("open hosts err", err)
		return
	}
	defer f.Close()
	r := bufio.NewReader(f)
	pos := 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("hosts read ok")
				break
			}
			fmt.Println("read hosts err", err)
			return
		}
		if strings.Contains(line, domain) {
			bytes := []byte(fmt.Sprintf("%s %s\n", ip, domain))
			f.WriteAt(bytes, int64(pos))
			return
		}
		pos += len(line)
	}
}
