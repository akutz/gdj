// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package canaries

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/govmomi/vim25/types"
)

var serializationTests = []struct {
	name   string
	file   string
	data   interface{}
	goType reflect.Type
}{
	{
		name:   "vminfo",
		file:   "./testdata/vminfo.json",
		data:   &vmInfoObjForTests,
		goType: reflect.TypeOf(types.VirtualMachineConfigInfo{}),
	},
	{
		name:   "retrieveResult",
		file:   "./testdata/retrieveResult.json",
		data:   &retrieveResultForTests,
		goType: reflect.TypeOf(types.RetrieveResult{}),
	},
}

func TestSerialization(t *testing.T) {
	for _, test := range serializationTests {
		t.Run(test.name+" Decode", func(t *testing.T) {
			f, err := os.Open(test.file)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			dec := NewGovmomiDecoder(f)

			data := reflect.New(test.goType).Interface()
			if err := dec.Decode(data); err != nil {
				t.Fatal(err)
			}

			a, e := data, test.data

			if diff := cmp.Diff(a, e); diff != "" {
				t.Errorf("mismatched %v: %s", test.name, diff)
			}
		})

		t.Run(test.name+" Encode", func(t *testing.T) {
			expJSON, err := os.ReadFile(test.file)
			if err != nil {
				t.Fatal(err)
			}

			var w bytes.Buffer
			_ = w
			enc := NewGovmomiEncoder(&w)

			if err := enc.Encode(test.data); err != nil {
				t.Fatal(err)
			}

			expected, actual := string(expJSON), w.String()
			assert.JSONEq(t, expected, actual)
		})
	}
}

var vmInfoObjForTests = types.VirtualMachineConfigInfo{
	ChangeVersion:         "2022-12-12T11:48:35.473645Z",
	Modified:              mustParseTime(time.RFC3339, "1970-01-01T00:00:00Z"),
	Name:                  "test",
	GuestFullName:         "VMware Photon OS (64-bit)",
	Version:               "vmx-20",
	Uuid:                  "422ca90b-853b-1101-3350-759f747730cc",
	CreateDate:            addrOfMustParseTime(time.RFC3339, "2022-12-12T11:47:24.685785Z"),
	InstanceUuid:          "502cc2a5-1f06-2890-6d70-ba2c55c5c2b7",
	NpivTemporaryDisabled: addrOfBool(true),
	LocationId:            "Earth",
	Template:              false,
	GuestId:               "vmwarePhoton64Guest",
	AlternateGuestName:    "",
	Annotation:            "Hello, world.",
	Files: types.VirtualMachineFileInfo{
		VmPathName:        "[datastore1] test/test.vmx",
		SnapshotDirectory: "[datastore1] test/",
		SuspendDirectory:  "[datastore1] test/",
		LogDirectory:      "[datastore1] test/",
	},
	Tools: &types.ToolsConfigInfo{
		ToolsVersion:            1,
		AfterPowerOn:            addrOfBool(true),
		AfterResume:             addrOfBool(true),
		BeforeGuestStandby:      addrOfBool(true),
		BeforeGuestShutdown:     addrOfBool(true),
		BeforeGuestReboot:       nil,
		ToolsUpgradePolicy:      "manual",
		SyncTimeWithHostAllowed: addrOfBool(true),
		SyncTimeWithHost:        addrOfBool(false),
		LastInstallInfo: &types.ToolsConfigInfoToolsLastInstallInfo{
			Counter: 0,
		},
	},
	Flags: types.VirtualMachineFlagInfo{
		EnableLogging:            addrOfBool(true),
		UseToe:                   addrOfBool(false),
		RunWithDebugInfo:         addrOfBool(false),
		MonitorType:              "release",
		HtSharing:                "any",
		SnapshotDisabled:         addrOfBool(false),
		SnapshotLocked:           addrOfBool(false),
		DiskUuidEnabled:          addrOfBool(false),
		SnapshotPowerOffBehavior: "powerOff",
		RecordReplayEnabled:      addrOfBool(false),
		FaultToleranceType:       "unset",
		CbrcCacheEnabled:         addrOfBool(false),
		VvtdEnabled:              addrOfBool(false),
		VbsEnabled:               addrOfBool(false),
	},
	DefaultPowerOps: types.VirtualMachineDefaultPowerOpInfo{
		PowerOffType:        "soft",
		SuspendType:         "hard",
		ResetType:           "soft",
		DefaultPowerOffType: "soft",
		DefaultSuspendType:  "hard",
		DefaultResetType:    "soft",
		StandbyAction:       "checkpoint",
	},
	RebootPowerOff: addrOfBool(false),
	Hardware: types.VirtualHardware{
		NumCPU:              1,
		NumCoresPerSocket:   1,
		AutoCoresPerSocket:  addrOfBool(true),
		MemoryMB:            2048,
		VirtualICH7MPresent: addrOfBool(false),
		VirtualSMCPresent:   addrOfBool(false),
		Device: []types.BaseVirtualDevice{
			&types.VirtualIDEController{
				VirtualController: types.VirtualController{
					VirtualDevice: types.VirtualDevice{
						Key: 200,
						DeviceInfo: &types.Description{
							Label:   "IDE 0",
							Summary: "IDE 0",
						},
					},
					BusNumber: 0,
				},
			},
			&types.VirtualIDEController{
				VirtualController: types.VirtualController{
					VirtualDevice: types.VirtualDevice{
						Key: 201,
						DeviceInfo: &types.Description{
							Label:   "IDE 1",
							Summary: "IDE 1",
						},
					},
					BusNumber: 1,
				},
			},
			&types.VirtualPS2Controller{
				VirtualController: types.VirtualController{
					VirtualDevice: types.VirtualDevice{
						Key: 300,
						DeviceInfo: &types.Description{
							Label:   "PS2 controller 0",
							Summary: "PS2 controller 0",
						},
					},
					BusNumber: 0,
					Device:    []int32{600, 700},
				},
			},
			&types.VirtualPCIController{
				VirtualController: types.VirtualController{
					VirtualDevice: types.VirtualDevice{
						Key: 100,
						DeviceInfo: &types.Description{
							Label:   "PCI controller 0",
							Summary: "PCI controller 0",
						},
					},
					BusNumber: 0,
					Device:    []int32{500, 12000, 14000, 1000, 15000, 4000},
				},
			},
			&types.VirtualSIOController{
				VirtualController: types.VirtualController{
					VirtualDevice: types.VirtualDevice{
						Key: 400,
						DeviceInfo: &types.Description{
							Label:   "SIO controller 0",
							Summary: "SIO controller 0",
						},
					},
					BusNumber: 0,
				},
			},
			&types.VirtualKeyboard{
				VirtualDevice: types.VirtualDevice{
					Key: 600,
					DeviceInfo: &types.Description{
						Label:   "Keyboard",
						Summary: "Keyboard",
					},
					ControllerKey: 300,
					UnitNumber:    addrOfInt32(0),
				},
			},
			&types.VirtualPointingDevice{
				VirtualDevice: types.VirtualDevice{
					Key:        700,
					DeviceInfo: &types.Description{Label: "Pointing device", Summary: "Pointing device; Device"},
					Backing: &types.VirtualPointingDeviceDeviceBackingInfo{
						VirtualDeviceDeviceBackingInfo: types.VirtualDeviceDeviceBackingInfo{
							UseAutoDetect: addrOfBool(false),
						},
						HostPointingDevice: "autodetect",
					},
					ControllerKey: 300,
					UnitNumber:    addrOfInt32(1),
				},
			},
			&types.VirtualMachineVideoCard{
				VirtualDevice: types.VirtualDevice{
					Key:           500,
					DeviceInfo:    &types.Description{Label: "Video card ", Summary: "Video card"},
					ControllerKey: 100,
					UnitNumber:    addrOfInt32(0),
				},
				VideoRamSizeInKB:       4096,
				NumDisplays:            1,
				UseAutoDetect:          addrOfBool(false),
				Enable3DSupport:        addrOfBool(false),
				Use3dRenderer:          "automatic",
				GraphicsMemorySizeInKB: 262144,
			},
			&types.VirtualMachineVMCIDevice{
				VirtualDevice: types.VirtualDevice{
					Key: 12000,
					DeviceInfo: &types.Description{
						Label: "VMCI device",
						Summary: "Device on the virtual machine PCI " +
							"bus that provides support for the " +
							"virtual machine communication interface",
					},
					ControllerKey: 100,
					UnitNumber:    addrOfInt32(17),
				},
				Id:                             -1,
				AllowUnrestrictedCommunication: addrOfBool(false),
				FilterEnable:                   addrOfBool(true),
			},
			&types.ParaVirtualSCSIController{
				VirtualSCSIController: types.VirtualSCSIController{
					VirtualController: types.VirtualController{
						VirtualDevice: types.VirtualDevice{
							Key: 1000,
							DeviceInfo: &types.Description{
								Label:   "SCSI controller 0",
								Summary: "VMware paravirtual SCSI",
							},
							ControllerKey: 100,
							UnitNumber:    addrOfInt32(3),
						},
						Device: []int32{2000},
					},
					HotAddRemove:       addrOfBool(true),
					SharedBus:          "noSharing",
					ScsiCtlrUnitNumber: 7,
				},
			},
			&types.VirtualAHCIController{
				VirtualSATAController: types.VirtualSATAController{
					VirtualController: types.VirtualController{
						VirtualDevice: types.VirtualDevice{
							Key: 15000,
							DeviceInfo: &types.Description{
								Label:   "SATA controller 0",
								Summary: "AHCI",
							},
							ControllerKey: 100,
							UnitNumber:    addrOfInt32(24),
						},
						Device: []int32{16000},
					},
				},
			},
			&types.VirtualCdrom{
				VirtualDevice: types.VirtualDevice{
					Key: 16000,
					DeviceInfo: &types.Description{
						Label:   "CD/DVD drive 1",
						Summary: "Remote device",
					},
					Backing: &types.VirtualCdromRemotePassthroughBackingInfo{
						VirtualDeviceRemoteDeviceBackingInfo: types.VirtualDeviceRemoteDeviceBackingInfo{
							UseAutoDetect: addrOfBool(false),
						},
					},
					Connectable:   &types.VirtualDeviceConnectInfo{AllowGuestControl: true, Status: "untried"},
					ControllerKey: 15000,
					UnitNumber:    addrOfInt32(0),
				},
			},
			&types.VirtualDisk{
				VirtualDevice: types.VirtualDevice{
					Key: 2000,
					DeviceInfo: &types.Description{
						Label:   "Hard disk 1",
						Summary: "4,194,304 KB",
					},
					Backing: &types.VirtualDiskFlatVer2BackingInfo{
						VirtualDeviceFileBackingInfo: types.VirtualDeviceFileBackingInfo{
							BackingObjectId: "1",
							FileName:        "[datastore1] test/test.vmdk",
							Datastore: &types.ManagedObjectReference{
								Type:  "Datastore",
								Value: "datastore-21",
							},
						},
						DiskMode:               "persistent",
						Split:                  addrOfBool(false),
						WriteThrough:           addrOfBool(false),
						ThinProvisioned:        addrOfBool(false),
						EagerlyScrub:           addrOfBool(false),
						Uuid:                   "6000C298-df15-fe89-ddcb-8ea33329595d",
						ContentId:              "e4e1a794c6307ce7906a3973fffffffe",
						ChangeId:               "",
						Parent:                 nil,
						DeltaDiskFormat:        "",
						DigestEnabled:          addrOfBool(false),
						DeltaGrainSize:         0,
						DeltaDiskFormatVariant: "",
						Sharing:                "sharingNone",
						KeyId:                  nil,
					},
					ControllerKey: 1000,
					UnitNumber:    addrOfInt32(0),
				},
				CapacityInKB:    4194304,
				CapacityInBytes: 4294967296,
				Shares:          &types.SharesInfo{Shares: 1000, Level: "normal"},
				StorageIOAllocation: &types.StorageIOAllocationInfo{
					Limit:       addrOfInt64(-1),
					Shares:      &types.SharesInfo{Shares: 1000, Level: "normal"},
					Reservation: addrOfInt32(0),
				},
				DiskObjectId:               "1-2000",
				NativeUnmanagedLinkedClone: addrOfBool(false),
			},
			&types.VirtualVmxnet3{
				VirtualVmxnet: types.VirtualVmxnet{
					VirtualEthernetCard: types.VirtualEthernetCard{
						VirtualDevice: types.VirtualDevice{
							Key: 4000,
							DeviceInfo: &types.Description{
								Label:   "Network adapter 1",
								Summary: "VM Network",
							},
							Backing: &types.VirtualEthernetCardNetworkBackingInfo{
								VirtualDeviceDeviceBackingInfo: types.VirtualDeviceDeviceBackingInfo{
									DeviceName:    "VM Network",
									UseAutoDetect: addrOfBool(false),
								},
								Network: &types.ManagedObjectReference{
									Type:  "Network",
									Value: "network-27",
								},
							},
							Connectable: &types.VirtualDeviceConnectInfo{
								MigrateConnect: "unset",
								StartConnected: true,
								Status:         "untried",
							},
							ControllerKey: 100,
							UnitNumber:    addrOfInt32(7),
						},
						AddressType:      "assigned",
						MacAddress:       "00:50:56:ac:4d:ed",
						WakeOnLanEnabled: addrOfBool(true),
						ResourceAllocation: &types.VirtualEthernetCardResourceAllocation{
							Reservation: addrOfInt64(0),
							Share: types.SharesInfo{
								Shares: 50,
								Level:  "normal",
							},
							Limit: addrOfInt64(-1),
						},
						UptCompatibilityEnabled: addrOfBool(true),
					},
				},
				Uptv2Enabled: addrOfBool(false),
			},
			&types.VirtualUSBXHCIController{
				VirtualController: types.VirtualController{
					VirtualDevice: types.VirtualDevice{
						Key: 14000,
						DeviceInfo: &types.Description{
							Label:   "USB xHCI controller ",
							Summary: "USB xHCI controller",
						},
						SlotInfo: &types.VirtualDevicePciBusSlotInfo{
							PciSlotNumber: -1,
						},
						ControllerKey: 100,
						UnitNumber:    addrOfInt32(23),
					},
				},

				AutoConnectDevices: addrOfBool(false),
			},
		},
		MotherboardLayout:   "i440bxHostBridge",
		SimultaneousThreads: 1,
	},
	CpuAllocation: &types.ResourceAllocationInfo{
		Reservation:           addrOfInt64(0),
		ExpandableReservation: addrOfBool(false),
		Limit:                 addrOfInt64(-1),
		Shares: &types.SharesInfo{
			Shares: 1000,
			Level:  types.SharesLevelNormal,
		},
	},
	MemoryAllocation: &types.ResourceAllocationInfo{
		Reservation:           addrOfInt64(0),
		ExpandableReservation: addrOfBool(false),
		Limit:                 addrOfInt64(-1),
		Shares: &types.SharesInfo{
			Shares: 20480,
			Level:  types.SharesLevelNormal,
		},
	},
	LatencySensitivity: &types.LatencySensitivity{
		Level: types.LatencySensitivitySensitivityLevelNormal,
	},
	MemoryHotAddEnabled: addrOfBool(false),
	CpuHotAddEnabled:    addrOfBool(false),
	CpuHotRemoveEnabled: addrOfBool(false),
	ExtraConfig: []types.BaseOptionValue{
		&types.OptionValue{Key: "nvram", Value: "test.nvram"},
		&types.OptionValue{Key: "svga.present", Value: "TRUE"},
		&types.OptionValue{Key: "pciBridge0.present", Value: "TRUE"},
		&types.OptionValue{Key: "pciBridge4.present", Value: "TRUE"},
		&types.OptionValue{Key: "pciBridge4.virtualDev", Value: "pcieRootPort"},
		&types.OptionValue{Key: "pciBridge4.functions", Value: "8"},
		&types.OptionValue{Key: "pciBridge5.present", Value: "TRUE"},
		&types.OptionValue{Key: "pciBridge5.virtualDev", Value: "pcieRootPort"},
		&types.OptionValue{Key: "pciBridge5.functions", Value: "8"},
		&types.OptionValue{Key: "pciBridge6.present", Value: "TRUE"},
		&types.OptionValue{Key: "pciBridge6.virtualDev", Value: "pcieRootPort"},
		&types.OptionValue{Key: "pciBridge6.functions", Value: "8"},
		&types.OptionValue{Key: "pciBridge7.present", Value: "TRUE"},
		&types.OptionValue{Key: "pciBridge7.virtualDev", Value: "pcieRootPort"},
		&types.OptionValue{Key: "pciBridge7.functions", Value: "8"},
		&types.OptionValue{Key: "hpet0.present", Value: "TRUE"},
		&types.OptionValue{Key: "RemoteDisplay.maxConnections", Value: "-1"},
		&types.OptionValue{Key: "sched.cpu.latencySensitivity", Value: "normal"},
		&types.OptionValue{Key: "vmware.tools.internalversion", Value: "0"},
		&types.OptionValue{Key: "vmware.tools.requiredversion", Value: "12352"},
		&types.OptionValue{Key: "migrate.hostLogState", Value: "none"},
		&types.OptionValue{Key: "migrate.migrationId", Value: "0"},
		&types.OptionValue{Key: "migrate.hostLog", Value: "test-36f94569.hlog"},
		&types.OptionValue{
			Key:   "viv.moid",
			Value: "c5b34aa9-d962-4a74-b7d2-b83ec683ba1b:vm-28:lIgQ2t7v24n2nl3N7K3m6IHW2OoPF4CFrJd5N+Tdfio=",
		},
	},
	DatastoreUrl: []types.VirtualMachineConfigInfoDatastoreUrlPair{
		{
			Name: "datastore1",
			Url:  "/vmfs/volumes/63970ed8-4abddd2a-62d7-02003f49c37d",
		},
	},
	SwapPlacement: "inherit",
	BootOptions: &types.VirtualMachineBootOptions{
		EnterBIOSSetup:       addrOfBool(false),
		EfiSecureBootEnabled: addrOfBool(false),
		BootDelay:            1,
		BootRetryEnabled:     addrOfBool(false),
		BootRetryDelay:       10000,
		NetworkBootProtocol:  "ipv4",
	},
	FtInfo:                       nil,
	RepConfig:                    nil,
	VAppConfig:                   nil,
	VAssertsEnabled:              addrOfBool(false),
	ChangeTrackingEnabled:        addrOfBool(false),
	Firmware:                     "bios",
	MaxMksConnections:            -1,
	GuestAutoLockEnabled:         addrOfBool(true),
	ManagedBy:                    nil,
	MemoryReservationLockedToMax: addrOfBool(false),
	InitialOverhead: &types.VirtualMachineConfigInfoOverheadInfo{
		InitialMemoryReservation: 214446080,
		InitialSwapReservation:   2541883392,
	},
	NestedHVEnabled: addrOfBool(false),
	VPMCEnabled:     addrOfBool(false),
	ScheduledHardwareUpgradeInfo: &types.ScheduledHardwareUpgradeInfo{
		UpgradePolicy:                  "never",
		ScheduledHardwareUpgradeStatus: "none",
	},
	ForkConfigInfo:         nil,
	VFlashCacheReservation: 0,
	VmxConfigChecksum: []uint8{
		0x69, 0xf7, 0xa7, 0x9e,
		0xd1, 0xc2, 0x21, 0x4b,
		0x6c, 0x20, 0x77, 0x0a,
		0x94, 0x94, 0x99, 0xee,
		0x17, 0x5d, 0xdd, 0xa3,
	},
	MessageBusTunnelEnabled: addrOfBool(false),
	GuestIntegrityInfo: &types.VirtualMachineGuestIntegrityInfo{
		Enabled: addrOfBool(false),
	},
	MigrateEncryption: "opportunistic",
	SgxInfo: &types.VirtualMachineSgxInfo{
		FlcMode:            "unlocked",
		RequireAttestation: addrOfBool(false),
	},
	ContentLibItemInfo:      nil,
	FtEncryptionMode:        "ftEncryptionOpportunistic",
	GuestMonitoringModeInfo: &types.VirtualMachineGuestMonitoringModeInfo{},
	SevEnabled:              addrOfBool(false),
	NumaInfo: &types.VirtualMachineVirtualNumaInfo{
		AutoCoresPerNumaNode:    addrOfBool(true),
		VnumaOnCpuHotaddExposed: addrOfBool(false),
	},
	PmemFailoverEnabled:          addrOfBool(false),
	VmxStatsCollectionEnabled:    addrOfBool(true),
	VmOpNotificationToAppEnabled: addrOfBool(false),
	VmOpNotificationTimeout:      -1,
	DeviceSwap: &types.VirtualMachineVirtualDeviceSwap{
		LsiToPvscsi: &types.VirtualMachineVirtualDeviceSwapDeviceSwapInfo{
			Enabled:    addrOfBool(true),
			Applicable: addrOfBool(false),
			Status:     "none",
		},
	},
	Pmem:         nil,
	DeviceGroups: &types.VirtualMachineVirtualDeviceGroups{},
}

var retrieveResultForTests = types.RetrieveResult{
	Token: "",
	Objects: []types.ObjectContent{

		{

			DynamicData: types.DynamicData{},
			Obj: types.ManagedObjectReference{

				Type:  "Folder",
				Value: "group-d1",
			},
			PropSet: []types.DynamicProperty{
				{

					Name: "alarmActionsEnabled",
					Val:  true,
				},
				{

					Name: "availableField",
					Val: types.ArrayOfCustomFieldDef{

						CustomFieldDef: []types.CustomFieldDef{},
					},
				},

				{

					Name: "childEntity",
					Val: types.ArrayOfManagedObjectReference{
						ManagedObjectReference: []types.ManagedObjectReference{},
					},
				},
				{
					Name: "childType",
					Val: types.ArrayOfString{
						String: []string{
							"Folder",
							"Datacenter"},
					},
				},
				{
					Name: "configIssue",
					Val: types.ArrayOfEvent{
						Event: []types.BaseEvent{},
					},
				},
				{
					Name: "configStatus",
					Val:  types.ManagedEntityStatusGray},
				{
					Name: "customValue",
					Val: types.ArrayOfCustomFieldValue{
						CustomFieldValue: []types.BaseCustomFieldValue{},
					},
				},
				{
					Name: "declaredAlarmState",
					Val: types.ArrayOfAlarmState{
						AlarmState: []types.AlarmState{
							{
								Key: "alarm-328.group-d1",
								Entity: types.ManagedObjectReference{
									Type:  "Folder",
									Value: "group-d1"},
								Alarm: types.ManagedObjectReference{
									Type:  "Alarm",
									Value: "alarm-328"},
								OverallStatus: "gray",
								Time:          time.Date(2023, time.January, 14, 8, 57, 35, 279575000, time.UTC),
								Acknowledged:  addrOfBool(false),
							},
							{
								Key: "alarm-327.group-d1",
								Entity: types.ManagedObjectReference{
									Type:  "Folder",
									Value: "group-d1"},
								Alarm: types.ManagedObjectReference{
									Type:  "Alarm",
									Value: "alarm-327"},
								OverallStatus: "green",
								Time:          time.Date(2023, time.January, 14, 8, 56, 40, 83607000, time.UTC),
								Acknowledged:  addrOfBool(false),
								EventKey:      756,
							},
							{
								DynamicData: types.DynamicData{},
								Key:         "alarm-326.group-d1",
								Entity: types.ManagedObjectReference{
									Type:  "Folder",
									Value: "group-d1"},
								Alarm: types.ManagedObjectReference{
									Type:  "Alarm",
									Value: "alarm-326"},
								OverallStatus: "green",
								Time: time.Date(2023,
									time.January,
									14,
									8,
									56,
									35,
									82616000,
									time.UTC),
								Acknowledged: addrOfBool(false),
								EventKey:     751,
							},
						},
					},
				},
				{
					Name: "disabledMethod",
					Val: types.ArrayOfString{
						String: []string{},
					},
				},
				{
					Name: "effectiveRole",
					Val: types.ArrayOfInt{
						Int: []int32{-1},
					},
				},
				{
					Name: "name",
					Val:  "Datacenters"},
				{
					Name: "overallStatus",
					Val:  types.ManagedEntityStatusGray},
				{
					Name: "permission",
					Val: types.ArrayOfPermission{
						Permission: []types.Permission{
							{
								Entity: &types.ManagedObjectReference{
									Value: "group-d1",
									Type:  "Folder",
								},
								Principal: "VSPHERE.LOCAL\\vmware-vsm-2bd917c6-e084-4d1f-988d-a68f7525cc94",
								Group:     false,
								RoleId:    1034,
								Propagate: true},
							{
								Entity: &types.ManagedObjectReference{
									Value: "group-d1",
									Type:  "Folder",
								},
								Principal: "VSPHERE.LOCAL\\topologysvc-2bd917c6-e084-4d1f-988d-a68f7525cc94",
								Group:     false,
								RoleId:    1024,
								Propagate: true},
							{
								Entity: &types.ManagedObjectReference{
									Value: "group-d1",
									Type:  "Folder",
								},
								Principal: "VSPHERE.LOCAL\\vpxd-extension-2bd917c6-e084-4d1f-988d-a68f7525cc94",
								Group:     false,
								RoleId:    -1,
								Propagate: true},
						},
					},
				},
				{
					Name: "recentTask",
					Val: types.ArrayOfManagedObjectReference{
						ManagedObjectReference: []types.ManagedObjectReference{
							{
								Type:  "Task",
								Value: "task-186"},
							{
								Type:  "Task",
								Value: "task-187"},
							{
								Type:  "Task",
								Value: "task-188"},
						},
					},
				},
				{
					Name: "tag",
					Val: types.ArrayOfTag{
						Tag: []types.Tag{},
					},
				},
				{
					Name: "triggeredAlarmState",
					Val: types.ArrayOfAlarmState{
						AlarmState: []types.AlarmState{},
					},
				},
				{
					Name: "value",
					Val: types.ArrayOfCustomFieldValue{
						CustomFieldValue: []types.BaseCustomFieldValue{},
					},
				},
			},
			MissingSet: nil,
		},
	},
}
