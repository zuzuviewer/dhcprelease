package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	frontIp = "01" +
		"01" +
		"06" +
		"00" +
		"13" +
		"db" +
		"0b" +
		"ce" +
		"0000" +
		"0000"
)

func parseIP(ip string) []byte {
	ret := make([]byte, 0, 4)
	ips := strings.Split(ip, ".")
	for _, v := range ips {
		value, err := strconv.Atoi(v)
		if err != nil {
			log.Panicf("parse ip %s ,v %s,failed,err %v\n", ip,v, err)
		}
		ret = append(ret, byte(value))
	}
	return ret
}

func parseMAC(mac string) []byte {
	macByte, err := net.ParseMAC(mac)
	if err != nil {
		log.Panicf("parse mac address %s failed,err %v\n", mac, err)
	}
	return macByte
}

func main() {
	if len(os.Args) <= 3 {
		log.Panicf("need client ip,mac address and dhcp server ip address")
	}
	clientIP, clientMAC, serverIP := parseIP(os.Args[1]), parseMAC(os.Args[2]), parseIP(os.Args[3])
	front, err := hex.DecodeString(frontIp)
	if err != nil {
		log.Panicf("parse front string failed,err %v\n", frontIp)
	}
	zeroIP := parseIP("0.0.0.0")
	releaseOrder := append(front, append(clientIP, append(zeroIP, append(zeroIP, zeroIP...)...)...)...)
	releaseOrder = append(releaseOrder, clientMAC...)
	log.Printf("len %d\n",len(releaseOrder))
	macAddressPadding, err := hex.DecodeString("00000000000000000000")
	if err != nil {
		log.Panicf("parse mac address padding failed,err %v\n", err)
	}
	releaseOrder = append(releaseOrder, macAddressPadding...)
	serverHostName, err := hex.DecodeString("00000000000000000000" +
		"00000000000000000000" +
		"00000000000000000000" +
		"00000000000000000000" +
		"00000000000000000000" +
		"00000000000000000000" +
		"00000000")
	if err != nil {
		log.Panicf("parse server host name failed,err %v\n", err)
	}
	releaseOrder = append(releaseOrder, serverHostName...)
	//append boot file name
	releaseOrder = append(releaseOrder, append(serverHostName, serverHostName...)...)
	magicCookie, err := hex.DecodeString("63825363")
	if err != nil {
		log.Panicf("parse magic cookie failed,err %v\n", err)
	}
	releaseOrder = append(releaseOrder, magicCookie...)
	messageType, err := hex.DecodeString("3501073604")
	if err != nil {
		log.Panicf("parse message type failed,err %v\n", err)
	}
	releaseOrder = append(releaseOrder, append(messageType, serverIP...)...)
	clientIdentifier, err := hex.DecodeString("3d0701")
	if err != nil {
		log.Panicf("parse clientIdentifier failed,err %v\n", err)
	}
	releaseOrder = append(releaseOrder, append(clientIdentifier, clientMAC...)...)
	end, err := hex.DecodeString("ff" +
		"00000000000000000000" +
		"00000000000000000000" +
		"00000000000000000000" +
		"0000000000000000000000")
	if err != nil {
		log.Panicf("parse end failed,err %v\n", err)
	}
	releaseOrder = append(releaseOrder, end...)
	raddr, err := net.ResolveUDPAddr("udp", os.Args[3]+fmt.Sprintf(":%d", 67))
	if err != nil {
		log.Panicf("resovle udp address failed,err %v'\n", err)
	}
	log.Printf("dhcp server %s\n,data len %d\n", raddr.String(),len(releaseOrder))
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Panicf("dial udp failed,err %v\n", err)
	}
	_, err = conn.Write(releaseOrder)
	if err != nil {
		log.Panicf("send release order to dhcp server failed,err %v\n", err)
	}
}
