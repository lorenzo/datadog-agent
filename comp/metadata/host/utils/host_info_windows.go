// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Portions of this code are taken from the gopsutil project
// https://github.com/shirou/gopsutil .  This code is licensed under the New BSD License
// copyright WAKAYAMA Shirou, and the gopsutil contributors

package utils

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/DataDog/datadog-agent/pkg/gohai/platform"
	"github.com/shirou/w32"
	"golang.org/x/sys/windows"

	"github.com/DataDog/datadog-agent/pkg/metadata/common"
	"github.com/DataDog/datadog-agent/pkg/util/cache"
	"github.com/DataDog/datadog-agent/pkg/util/winutil"
)

var (
	modkernel          = windows.NewLazyDLL("kernel32.dll")
	procGetTickCount64 = modkernel.NewProc("GetTickCount64")
)

// InfoStat describes the host status.  This is not in the psutil but it useful.
type InfoStat struct {
	Hostname             string `json:"hostname"`
	Uptime               uint64 `json:"uptime"`
	BootTime             uint64 `json:"bootTime"`
	Procs                uint64 `json:"procs"`           // number of processes
	OS                   string `json:"os"`              // ex: freebsd, linux
	Platform             string `json:"platform"`        // ex: ubuntu, linuxmint
	PlatformFamily       string `json:"platformFamily"`  // ex: debian, rhel
	PlatformVersion      string `json:"platformVersion"` // version of the complete OS
	KernelVersion        string `json:"kernelVersion"`   // version of the OS kernel (if available)
	KernelArch           string `json:"kernelArch"`
	VirtualizationSystem string `json:"virtualizationSystem"`
	VirtualizationRole   string `json:"virtualizationRole"` // guest or host
	HostID               string `json:"hostid"`             // ex: uuid
}

// GetInformation returns an InfoStat object, filled in with various operating system metadata
func GetInformation() *InfoStat {
	info, _ := cache.Get[*InfoStat](
		cacheKey,
		func() (*InfoStat, error) {
			info := &InfoStat{}
			info.Hostname, _ = os.Hostname()

			upTime := time.Duration(getTickCount64()) * time.Millisecond
			bootTime := time.Now().Add(-upTime)
			info.Uptime = uint64(upTime.Seconds())
			info.BootTime = uint64(bootTime.Unix())
			pids, _ := Pids()
			info.Procs = uint64(len(pids))
			info.OS = runtime.GOOS

			info.KernelArch = runtime.GOARCH

			pi := platform.CollectInfo()
			info.Platform, _ = pi.OS.Value()
			info.PlatformFamily, _ = pi.OS.Value()

			info.PlatformVersion, _ = winutil.GetWindowsBuildString()
			info.HostID = common.GetUUID()
			return info, nil
		})
	return info
}

////////////////////////////////////////////////////////////
// windows helpers
//

// getTickCount64() returns the time, in milliseconds, that have elapsed since
// the system was started
func getTickCount64() int64 {
	ret, _, _ := procGetTickCount64.Call()
	return int64(ret)
}

// Pids returns a list of process ids.
func Pids() ([]int32, error) {

	// inspired by https://gist.github.com/henkman/3083408
	// and https://github.com/giampaolo/psutil/blob/1c3a15f637521ba5c0031283da39c733fda53e4c/psutil/arch/windows/process_info.c#L315-L329
	var ret []int32
	var read uint32
	var psSize uint32 = 1024
	const dwordSize uint32 = 4

	for {
		ps := make([]uint32, psSize)
		if !w32.EnumProcesses(ps, uint32(len(ps)), &read) {
			return nil, fmt.Errorf("could not get w32.EnumProcesses")
		}
		if uint32(len(ps)) == read { // ps buffer was too small to host every results, retry with a bigger one
			psSize += 1024
			continue
		}
		for _, pid := range ps[:read/dwordSize] {
			ret = append(ret, int32(pid))
		}
		return ret, nil

	}
}
