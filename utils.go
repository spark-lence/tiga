package tiga

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func GetNowDatetime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
func GetNowDate() string {
	return time.Now().Format("2006-01-02")
}
func GetMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
func Map2String(data map[string]interface{}) (string, error) {
	str, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("marshal with error: %+v\n", err)
		return "", err
	}

	return string(str), nil
}
func ThousandsSeparatorToInt(src string) int {
	numStr := strings.Replace(src, ",", "", -1)

	// 将结果转换为 int
	num, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}
	return num

}
func Bytes2Base64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
func Bytes2BaseURL64(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}
func Base64URL2String(data string) (string, error) {
	bytes, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func Base64URL2Bytes(data string) ([]byte, error) {
	bytes, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
func Base642Bytes(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}
func URLTemplateRender(urlTemplate string, params map[string]interface{}) (string, error) {
	// 创建一个新的模板
	tmpl, err := template.New("url").Parse(urlTemplate)
	if err != nil {
		return "", fmt.Errorf("parse url template error: %w", err)
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, params)
	if err != nil {
		return "", fmt.Errorf("execute url template error: %w", err)
	}

	// 打印或使用结果
	renderedURL := buf.String()
	return renderedURL, nil
}

func InterfaceToBytes(data interface{}) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
func GetLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("Failed to get interfaces: %w", err)
	}
	ips := make([]string, 0)
	for _, intf := range interfaces {
		// 跳过非活动的网络接口和环回接口
		if intf.Flags&net.FlagUp == 0 || intf.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取网络接口的地址
		addrs, err := intf.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// 检查 IP 地址是否为 IPv4 地址
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // 非 IPv4 地址
			}
			ips = append(ips, ip.String())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ips)))
	if len(ips) > 0 {
		return ips[0], nil
	}
	return "", fmt.Errorf("Failed to get local IP address")
}

func GetMac() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("Failed to get interfaces: %v", err)
	}
	macs := make([]string, 0)
	for _, intf := range interfaces {
		// 跳过环回接口和没有 MAC 地址的接口
		if intf.Flags&net.FlagLoopback != 0 || intf.HardwareAddr == nil {
			continue
		}
		macs = append(macs, intf.HardwareAddr.String())
		// return intf.HardwareAddr.String(), nil
	}
	if len(macs) > 0 {
		sort.Strings(macs)
		return strings.Join(macs, "|"), nil
	}
	return "", fmt.Errorf("Failed to get mac address")
}

func GetMACAddress(ipAddr string) (string, error) {
	// 将字符串格式的IP地址解析为net.IP
	ip := net.ParseIP(ipAddr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address")
	}

	// 获取所有的网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	// 遍历所有网络接口
	for _, iface := range interfaces {
		// 获取每个接口的地址列表
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		// 检查IP地址是否属于当前接口
		for _, addr := range addrs {
			var networkIP net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				networkIP = v.IP
			case *net.IPAddr:
				networkIP = v.IP
			}

			// 如果IP匹配，返回该接口的MAC地址
			if ip.Equal(networkIP) {
				return iface.HardwareAddr.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no interface found for IP address %s", ipAddr)
}
func ArrayContainsString(arr []string, str string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}

func IntToBytes(number int) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, number)
	if err != nil {
		return nil, fmt.Errorf("binary.Write failed:%w", err)
	}
	return buf.Bytes(), nil
}
