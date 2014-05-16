package util

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
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

type MEMORYSTATUSEX struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

type PDH_FMT_COUNTERVALUE_DOUBLE struct {
	CStatus     uint32
	DoubleValue float64
}

type PDH_FMT_COUNTERVALUE_ITEM_DOUBLE struct {
	Name     *uint16
	FmtValue PDH_FMT_COUNTERVALUE_DOUBLE
}

const (
	ERROR_SUCCESS      = 0
	DRIVE_REMOVABLE    = 2
	DRIVE_FIXED        = 3
	HKEY_LOCAL_MACHINE = 0x80000002
	RRF_RT_REG_SZ      = 0x00000002
	RRF_RT_REG_DWORD   = 0x00000010
	PDH_FMT_DOUBLE     = 0x00000200
	PDH_INVALID_DATA   = 0xc0000bc6
)

var (
	modadvapi32 = syscall.NewLazyDLL("advapi32.dll")
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	modpdh      = syscall.NewLazyDLL("pdh.dll")

	RegGetValue                 = modadvapi32.NewProc("RegGetValueW")
	GetSystemInfo               = modkernel32.NewProc("GetSystemInfo")
	GetTickCount                = modkernel32.NewProc("GetTickCount")
	GetDiskFreeSpaceEx          = modkernel32.NewProc("GetDiskFreeSpaceExW")
	GetLogicalDriveStrings      = modkernel32.NewProc("GetLogicalDriveStringsW")
	GetDriveType                = modkernel32.NewProc("GetDriveTypeW")
	QueryDosDevice              = modkernel32.NewProc("QueryDosDeviceW")
	GetVolumeInformationW       = modkernel32.NewProc("GetVolumeInformationW")
	GlobalMemoryStatusEx        = modkernel32.NewProc("GlobalMemoryStatusEx")
	PdhOpenQuery                = modpdh.NewProc("PdhOpenQuery")
	PdhAddCounter               = modpdh.NewProc("PdhAddCounterW")
	PdhCollectQueryData         = modpdh.NewProc("PdhCollectQueryData")
	PdhGetFormattedCounterValue = modpdh.NewProc("PdhGetFormattedCounterValue")
	PdhCloseQuery               = modpdh.NewProc("PdhCloseQuery")
)

func RegGetInt(hKey uint32, subKey string, value string) (uint32, error) {
	var num, numlen uint32
	numlen = 4
	ret, _, err := RegGetValue.Call(
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

func RegGetString(hKey uint32, subKey string, value string) (string, error) {
	var bufLen uint32
	RegGetValue.Call(
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
	ret, _, err := RegGetValue.Call(
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

type CounterInfo struct {
	PostName    string
	CounterName string
	Counter     syscall.Handle
}

func CreateQuery() (syscall.Handle, error) {
	var query syscall.Handle
	r, _, err := PdhOpenQuery.Call(0, 0, uintptr(unsafe.Pointer(&query)))
	if r != 0 {
		return 0, err
	}
	return query, nil
}

func CreateCounter(query syscall.Handle, k, v string) (*CounterInfo, error) {
	var counter syscall.Handle
	r, _, err := PdhAddCounter.Call(
		uintptr(query),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(v))),
		0,
		uintptr(unsafe.Pointer(&counter)))
	if r != 0 {
		return nil, err
	}
	return &CounterInfo{
		PostName:    k,
		CounterName: v,
		Counter:     counter,
	}, nil
}

func GetAdapterList() (*syscall.IpAdapterInfo, error) {
	b := make([]byte, 1000)
	l := uint32(len(b))
	a := (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
	err := syscall.GetAdaptersInfo(a, &l)
	if err == syscall.ERROR_BUFFER_OVERFLOW {
		b = make([]byte, l)
		a = (*syscall.IpAdapterInfo)(unsafe.Pointer(&b[0]))
		err = syscall.GetAdaptersInfo(a, &l)
	}
	if err != nil {
		return nil, os.NewSyscallError("GetAdaptersInfo", err)
	}
	return a, nil
}

func BytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}
