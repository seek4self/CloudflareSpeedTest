// update v2ray config
package task

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const configFile = "./config.json"

var ConfigV2Ray = false

type v2RayConfig struct {
	Policy    interface{} `json:"policy"`
	Log       interface{} `json:"log"`
	InBounds  interface{} `json:"inbounds"`
	OutBounds []struct {
		Tag      string `json:"tag"`
		Protocol string `json:"protocol"`
		Settings struct {
			Vnext []struct {
				Address string      `json:"address"`
				Port    int         `json:"port"`
				Users   interface{} `json:"users"`
			} `json:"vnext"`
		} `json:"settings"`
	} `json:"outbounds"`
	Stats   interface{} `json:"stats"`
	API     interface{} `json:"api"`
	DNS     interface{} `json:"dns"`
	Routing interface{} `json:"routing"`
}

func replaceAddr(addr string) {
	config, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println("read v2ray config.json err", err)
		return
	}
	v2ray := v2RayConfig{}
	if err := json.Unmarshal(config, &v2ray); err != nil {
		fmt.Println("Unmarshal v2ray config.json err", err)
		return
	}
	v2ray.OutBounds[0].Settings.Vnext[0].Address = addr
	newConf, err := json.Marshal(v2ray)
	if err != nil {
		fmt.Println("Marshal v2ray config.json err", err)
		return
	}
	out := bytes.Buffer{}
	_ = json.Indent(&out, newConf, "", "\t")
	err = ioutil.WriteFile("./config.json", out.Bytes(), 0644)
	if err != nil {
		fmt.Println("Write v2ray config.json err", err)
		return
	}
	return
}
