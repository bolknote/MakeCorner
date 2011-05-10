package ini

import (
    "os"
    "bufio"
    "strings"
)

type inisection map[string]string
type inifile map[string]inisection

func ParseFile(name string) inifile {
    f, err := os.OpenFile(name, os.O_RDONLY, 0666)
    if err != nil {
        return nil
    }

    defer f.Close()

    r := bufio.NewReader(f)

    one := inisection{}             // одна, текущая секция
    ini := inifile   {}             // все секции
    section := ""                   // текущее название секции

    for {
        line, err := r.ReadString('\n')

        if err != nil {
            break
        }

        chunks := strings.Split(strings.TrimRight(line, "\r\n"), "=", 3)
        if len(chunks) == 2 {
            key, value := strings.Trim(chunks[0], "\t "), strings.Trim(chunks[1], "\t ")

            last := len(value) - 1
            if last > 0 && value[0] == '"' && value[last] == '"' {
                value = string(value[1:last])
            }

            one[key] = string(value)
        } else {
            last := len(chunks[0]) - 1

            if chunks[0][0] == '[' && chunks[0][last] == ']' {
                ini[section] = one
                one = inisection{}
                section = string(chunks[0][1:last])
            }
        }
    }

    // «поджимаем хвост», поскольку предыдущая секция попадает в map
    // только когда появляется следующая, последнюю нужно перекладывать «руками»
    if len(one) > 0 {
        ini[section] = one
    }

    return ini
}
