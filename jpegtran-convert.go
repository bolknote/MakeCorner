package main

import (
    "os"
    "bufio"
    "encoding/git85"
    "strings"
    "syscall"
)

func main() {
	jpegtran := "jpegtran." + syscall.OS + ".bz2"

	fr, _ := os.OpenFile(jpegtran, os.O_RDONLY, 0666)
	defer fr.Close()
	r := bufio.NewReader(fr)

	info, _ := fr.Stat()
	src := make([]byte, int(info.Size))
	r.Read(src)

	var dst []byte = make([]byte, git85.EncodedLen(len(src)))
	git85.Encode(dst, src)

	fw, _ := os.OpenFile("jpegtran.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	defer fw.Close()

	w := bufio.NewWriter(fw)
	str := (string)(dst)
	str = strings.Replace(str, "\\", `\\`, -1)
	str = strings.Replace(str, "\n", "\\n", -1)

    w.WriteString("package jpegtran\n\nconst Jpegtran = \"")
	w.WriteString(str)
	w.WriteString("\"\n")
	w.Flush()
}
