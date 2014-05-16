package spec

import (
	"net"
	"os"
	"syscall"
	"unsafe"
)

func bytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}

// getAdapterList return list of adapter information
func getAdapterList() (*syscall.IpAdapterInfo, error) {
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

func (g *InterfaceGenerator) Generate() (interface{}, error) {
	results := make([]map[string]interface{}, 0)

	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ai, err := getAdapterList()
	if err != nil {
		return nil, err
	}

	for _, ifi := range ifs {
		addr, err := ifi.Addrs()
		if err != nil {
			return nil, err
		}
		name := ifi.Name
		for ; ai != nil; ai = ai.Next {
			if ifi.Index == int(ai.Index) {
				name = bytePtrToString(&ai.Description[0])
			}
		}

		results = append(results, map[string]interface{}{
			"name":       name,
			"ipAddress":  addr[0].String(),
			"macAddress": ifi.HardwareAddr.String(),
		})
	}

	return results, nil
}
