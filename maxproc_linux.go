package main

import "runtime"
import "os"
import "bufio"
import "strings"
import "strconv"

// выставляем количество потоков = количеству процессоров
func setmaxprocs() {
    const cpuinfo = "/Users/bolk/Проекты/Corner/cpuinfo"

    if f, e := os.OpenFile(cpuinfo, os.O_RDONLY, 0666); e == nil {
        defer f.Close()

        r := bufio.NewReader(f)
        for {
            line, e := r.ReadString('\n')

            if e != nil {
                break
            }

            chunks := strings.Split(strings.TrimRight(line, "\r\n"), ":", 2)

            if len(chunks) == 2 {
                key, value := strings.Trim(chunks[0], "\t "), strings.Trim(chunks[1], "\t ")

                if key == "cpu cores" {
                    if n, _ := strconv.Atoi(value); n > 0 {
                        runtime.GOMAXPROCS(int(n))
                    }

                    return
                }
            }
        }
    }
}
