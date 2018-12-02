// +build ignore

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
)

func main() {
	p, _ := ioutil.ReadFile("ttf/arialuni.ttf")
	buf := bytes.NewBuffer(nil)
	w := gzip.NewWriter(buf)
	w.Write(p)
	w.Close()

	fmt.Println("package main")
	fmt.Println()

	fmt.Println("import (")
	fmt.Println("\t\"bytes\"")
	fmt.Println("\t\"compress/gzip\"")
	fmt.Println("\t\"encoding/base64\"")
	fmt.Println("\t\"io/ioutil\"")
	fmt.Println(")")
	fmt.Println()

	fmt.Println("var TTF []byte")
	fmt.Println()

	fmt.Println("func init() {")
	fmt.Println("\tp, err := base64.StdEncoding.DecodeString(ttf)")
	fmt.Println("\tif err != nil {")
	fmt.Println("\t\tpanic(err)")
	fmt.Println("\t}")
	fmt.Println("\tbr := bytes.NewReader(p)")
	fmt.Println("\tr, err := gzip.NewReader(br)")
	fmt.Println("\tif err != nil {")
	fmt.Println("\t\tpanic(err)")
	fmt.Println("\t}")
	fmt.Println("\tb, err := ioutil.ReadAll(r)")
	fmt.Println("\tif err != nil {")
	fmt.Println("\t\tpanic(err)")
	fmt.Println("\t}")
	fmt.Println("\tr.Close()")
	fmt.Println("\tTTF = b")
	fmt.Println("}")
	fmt.Println()

	fmt.Printf("var ttf = `")
	fmt.Print(base64.StdEncoding.EncodeToString(buf.Bytes()))
	fmt.Println("`")
}
