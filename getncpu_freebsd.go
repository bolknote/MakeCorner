package main
import "syscall"

func getncpu() int {
    if n, _ := syscall.SysctlUint32("hw.ncpu"); n > 0 {
        return int(n)
    }
}
