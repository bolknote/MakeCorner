package main

import "os"
import "bufio"
import "encoding/git85"
import "fmt"
import "strings"

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

	fw, e := os.OpenFile("jt.go", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	fmt.Println(e)
	defer fw.Close()

	w := bufio.NewWriter(fw)
	str := (string)(dst)
	//fmt.Println(dst)
	str = strings.Replace(str, "\\", `\\`, -1)
	str = strings.Replace(str, "\n", "\\n", -1)
	fmt.Println(len(dst), len(str))

	w.WriteString("\tjpegtran := \"")
	w.WriteString(str)
	w.WriteString("\"\n")
	w.Flush()
}
