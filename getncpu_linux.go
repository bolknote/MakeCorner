package ncpu

import "os"
import "bufio"
import "strings"

func Getncpu() int {
    const cpuinfo = "/proc/cpuinfo"
    n := 0

    if f, e := os.OpenFile(cpuinfo, os.O_RDONLY, 0666); e == nil {
        defer f.Close()

        r := bufio.NewReader(f)
        for {
            line, e := r.ReadString('\n')

            if e != nil {
                break
            }

            chunks := strings.Split(strings.TrimRight(line, "\r\n"), ":", 2)

            if len(chunks) == 2 && strings.Trim(chunks[0], " \t") == "processor" {
                n++
            }
        }
    }

    return n
}
