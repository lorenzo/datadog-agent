// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build linux

package probes

import manager "github.com/DataDog/ebpf-manager"

// linkProbes holds the list of probes used to track link events
var linkProbes = []*manager.Probe{
	{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID:          SecurityAgentUID,
			EBPFFuncName: "hook_vfs_link",
		},
	},
	{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID:          SecurityAgentUID,
			EBPFFuncName: "hook_do_linkat",
		},
	},
	{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID:          SecurityAgentUID,
			EBPFFuncName: "rethook_do_linkat",
		},
	},
}

func getLinkProbe(fentry bool) []*manager.Probe {
	linkProbes = append(linkProbes, ExpandSyscallProbes(&manager.Probe{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID: SecurityAgentUID,
		},
		SyscallFuncName: "link",
	}, fentry, EntryAndExit|SupportFentry|SupportFexit)...)
	linkProbes = append(linkProbes, ExpandSyscallProbes(&manager.Probe{
		ProbeIdentificationPair: manager.ProbeIdentificationPair{
			UID: SecurityAgentUID,
		},
		SyscallFuncName: "linkat",
	}, fentry, EntryAndExit|SupportFentry|SupportFexit)...)
	return linkProbes
}
