package util

import (
	"strings"
)

type LightningAddr struct {
	Pubkey string
	Host   string
}

func NewLightningAddr(lightningAddr string) *LightningAddr {
	splitLightningAddr := strings.Split(lightningAddr, "@")

	if len(splitLightningAddr) == 2 {
		return &LightningAddr{
			Pubkey: splitLightningAddr[0],
			Host:   splitLightningAddr[1],
		}
	}

	return &LightningAddr{
		Pubkey: lightningAddr,
	}
}

func (l *LightningAddr) Hostname() string {
	hostSplit := strings.Split(l.Host, ":")

	if len(hostSplit) == 2 {
		return hostSplit[0]
	}

	return l.Host
}
