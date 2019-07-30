package main

import (
	"log"
    "io"
    "io/ioutil"
    "os"
    "strings"
)

// Taken from https://stackoverflow.com/questions/17796043/how-to-embed-files-into-golang-binaries

// Reads all .html files in the static folder
// and encodes them as strings literals in html.go
func main() {
    fs, _ := ioutil.ReadDir("./static")
    out, _ := os.Create("textfiles.go")
    out.Write([]byte("package main \n\nconst (\n"))
    for _, f := range fs {
        if strings.HasSuffix(f.Name(), ".html") {
            name := strings.TrimSuffix(f.Name(), ".html")
            out.Write([]byte(name + "HTML = `"))
            f, _ := os.Open("static/" + f.Name())
            io.Copy(out, f)
            out.Write([]byte("`\n"))
            log.Print("Embedded " + f.Name() + " in the go binary as " + name + "HTML")
        }
    }
    out.Write([]byte(")\n"))
    out.Close()
}