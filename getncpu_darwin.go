package getncpu
import "syscall"

func Getncpu() int {
    n, _ := syscall.SysctlUint32("hw.ncpu")
    return int(n)
}
