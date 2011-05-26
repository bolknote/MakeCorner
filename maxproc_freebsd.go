package main

import "syscall"
import "runtime"

// выставляем количество потоков = количествую процессоров
func setmaxprocs() {
    if n, _ := syscall.SysctlUint32("hw.ncpu"); n > 0 {
        runtime.GOMAXPROCS(int(n))
    }
}
