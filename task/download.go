package task

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"time"

	"CloudflareSpeedTest/utils"
)

const (
	bufferSize                     = 1 << 20 // MB
	defaultURL                     = "https://cf.xiu2.xyz/Github/CloudflareSpeedTest.png"
	defaultTimeout                 = 10 * time.Second
	defaultDisableDownload         = false
	defaultTestNum                 = 10
	defaultMinSpeed        float64 = 0.0
)

var (
	// download test url
	URL = defaultURL
	// download timeout
	Timeout = defaultTimeout
	// disable download
	Disable = defaultDisableDownload

	TestCount = defaultTestNum
	MinSpeed  = defaultMinSpeed
)

func checkDownloadDefault() {
	if URL == "" {
		URL = defaultURL
	}
	if Timeout <= 0 {
		Timeout = defaultTimeout
	}
	if TestCount <= 0 {
		TestCount = defaultTestNum
	}
	if MinSpeed <= 0.0 {
		MinSpeed = defaultMinSpeed
	}
}

func TestDownloadSpeed(ipSet utils.PingDelaySet) (speedSet utils.DownloadSpeedSet) {
	checkDownloadDefault()
	if Disable {
		return utils.DownloadSpeedSet(ipSet)
	}
	if len(ipSet) <= 0 { // IP数组长度(IP数量) 大于 0 时才会继续下载测速
		fmt.Println("\n[信息] 延迟测速结果 IP 数量为 0，跳过下载测速。")
		return
	}
	testNum := TestCount
	if len(ipSet) < TestCount || MinSpeed > 0 { // 如果IP数组长度(IP数量) 小于下载测速数量（-dn），则次数修正为IP数
		testNum = len(ipSet)
	}

	fmt.Printf("开始下载测速（下载速度下限：%.2f MB/s，下载测速数量：%d，下载测速队列：%d）：\n", MinSpeed, TestCount, testNum)
	bar := utils.NewBar(TestCount)
	for i := 0; i < testNum; i++ {
		speed := downloadHandler(ipSet[i].IP)
		ipSet[i].DownloadSpeed = speed
		// 在每个 IP 下载测速后，以 [下载速度下限] 条件过滤结果
		if speed >= MinSpeed*(1<<20) {
			bar.Grow(1)
			speedSet = append(speedSet, ipSet[i]) // 高于下载速度下限时，添加到新数组中
			if len(speedSet) == TestCount {       // 凑够满足条件的 IP 时（下载测速数量 -dn），就跳出循环
				break
			}
		}
	}
	bar.Done()
	if len(speedSet) == 0 { // 没有符合速度限制的数据，返回所有测试数据
		speedSet = utils.DownloadSpeedSet(ipSet)
	}
	// 按速度排序
	sort.Sort(speedSet)
	return
}

func getDialContext(ip *net.IPAddr) func(ctx context.Context, network, address string) (net.Conn, error) {
	fakeSourceAddr := fmt.Sprintf("%s:443", ip.String())
	if IPv6 { // IPv6 需要加上 []
		fakeSourceAddr = fmt.Sprintf("[%s]:443", ip.String())
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, fakeSourceAddr)
	}
}

// return download Speed
func downloadHandler(ip *net.IPAddr) float64 {
	client := &http.Client{
		Transport: &http.Transport{DialContext: getDialContext(ip)},
		Timeout:   Timeout,
	}
	// fmt.Println("test", ip.String())
	response, err := client.Get(URL)
	if err != nil {
		return 0.0
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return 0.0
	}
	timer := time.NewTimer(Timeout)
	size, cost := int64(0), time.Duration(0)
	for i := int64(0); i < response.ContentLength/bufferSize/2; i++ {
		select {
		case <-timer.C:
			break
		default:
		}
		start := time.Now()
		n, err := io.CopyN(io.Discard, response.Body, bufferSize)
		if err != nil {
			// fmt.Println("read body err", err, ip.String())
			break
		}
		// fmt.Println("spent time", time.Since(start).Seconds(), n)
		size += n
		cost += time.Since(start)
	}
	return float64(size) / cost.Seconds()
}
