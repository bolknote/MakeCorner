package main
import "syscall"

func getncpu() int {
    if n, _ := strconv.Atoi(os.Getenv("NUMBER_OF_PROCESSORS"); n > 0 {
        return n
    }
}
