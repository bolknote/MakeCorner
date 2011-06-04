package getncpu

import "syscall"

func Getncpu() int {
	n, _ := strconv.Atoi(os.Getenv("NUMBER_OF_PROCESSORS"))
	return n
}
