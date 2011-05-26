package main
import "syscall"

func getncpu() int {
    n, _ := strconv.Atoi(os.Getenv("NUMBER_OF_PROCESSORS")
    return n
}
