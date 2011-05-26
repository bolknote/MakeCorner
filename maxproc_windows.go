package main

import "syscall"
import "runtime"

// выставляем количество потоков = количествую процессоров
func setmaxprocs() {
    if n, _ := strconv.Atoi(os.Getenv("NUMBER_OF_PROCESSORS"); n > 0 {
        runtime.GOMAXPROCS(n)
    }
}
