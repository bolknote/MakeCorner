package main

import "os"
import "bufio"
import "encoding/git85"
//import "fmt"

func main() {
    jpegtran := "jpegtran.bz2"

    fr, _ := os.OpenFile(jpegtran, os.O_RDONLY, 0666)
    defer fr.Close()
    r := bufio.NewReader(fr)


    info, _ := fr.Stat()
    src := make([]byte, int(info.Size))
    r.Read(src)

    var dst []byte = make([]byte, git85.EncodedLen(len(src)))
    git85.Encode(dst, src)

    fw, _ := os.OpenFile("jt.go", os.O_RDWR, 0666)
    defer fw.Close()

    w := bufio.NewWriter(fw)

    w.WriteString((string)(dst))
}
