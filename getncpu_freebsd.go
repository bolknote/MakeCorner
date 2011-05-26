package main
import "syscall"

func getncpu() int {
    n, _ := syscall.SysctlUint32("hw.ncpu")
    return int(n)
}
