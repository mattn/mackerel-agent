package spec

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	ERROR_SUCCESS      = 0
	DRIVE_FIXED        = 3
	HKEY_LOCAL_MACHINE = 0x80000002
	RRF_RT_REG_SZ      = 0x00000002
	RRF_RT_REG_DWORD   = 0x00000010
	PDH_FMT_DOUBLE     = 0x00000200
	PDH_INVALID_DATA   = 0xc0000bc6
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	modadvapi32 = syscall.NewLazyDLL("advapi32.dll")

	procRegGetValue   = modadvapi32.NewProc("RegGetValueW")
	procGetSystemInfo = modkernel32.NewProc("GetSystemInfo")
)

type SYSTEM_INFO struct {
	ProcessorArchitecture     uint16
	PageSize                  uint32
	MinimumApplicationAddress *byte
	MaximumApplicationAddress *byte
	ActiveProcessorMask       *byte
	NumberOfProcessors        uint32
	ProcessorType             uint32
	AllocationGranularity     uint32
	ProcessorLevel            uint16
	ProcessorRevision         uint16
}

func regGetString(hKey uint32, subKey string, value string) (string, error) {
	var bufLen uint32
	procRegGetValue.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(subKey))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(value))),
		uintptr(RRF_RT_REG_SZ),
		0,
		0,
		uintptr(unsafe.Pointer(&bufLen)))
	if bufLen == 0 {
		return "", errors.New("Can't get size of registry value")
	}

	buf := make([]uint16, bufLen)
	ret, _, err := procRegGetValue.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(subKey))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(value))),
		uintptr(RRF_RT_REG_SZ),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufLen)))
	if ret != ERROR_SUCCESS {
		return "", err
	}

	return syscall.UTF16ToString(buf), nil
}

func regGetInt(hKey uint32, subKey string, value string) (uint32, error) {
	var num, numlen uint32
	numlen = 4
	ret, _, err := procRegGetValue.Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(subKey))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(value))),
		uintptr(RRF_RT_REG_DWORD),
		0,
		uintptr(unsafe.Pointer(&num)),
		uintptr(unsafe.Pointer(&numlen)))
	if ret != ERROR_SUCCESS {
		return 0, err
	}

	return num, nil
}

func (g *CPUGenerator) Generate() (interface{}, error) {
	results := make([]map[string]interface{}, 0)

	var systemInfo SYSTEM_INFO
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&systemInfo)))

	for i := uint32(0); i < systemInfo.NumberOfProcessors; i++ {
		processorName, err := regGetString(
			HKEY_LOCAL_MACHINE,
			fmt.Sprintf(`HARDWARE\DESCRIPTION\System\CentralProcessor\%d`, i),
			`ProcessorNameString`)
		if err != nil {
			return nil, err
		}
		processorMHz, err := regGetInt(
			HKEY_LOCAL_MACHINE,
			fmt.Sprintf(`HARDWARE\DESCRIPTION\System\CentralProcessor\%d`, i),
			`~MHz`)
		if err != nil {
			return nil, err
		}
		vendorIdentifier, err := regGetString(
			HKEY_LOCAL_MACHINE,
			fmt.Sprintf(`HARDWARE\DESCRIPTION\System\CentralProcessor\%d`, i),
			`VendorIdentifier`)
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]interface{}{
			"model_name": processorName,
			"mhz":        processorMHz,
			"model":      systemInfo.ProcessorArchitecture,
			"vendor_id":  vendorIdentifier,
		})
	}
	return results, nil
}
