package main

import (
	"flag"
	"fmt"
	"html"
	"net/http"
	"os"
	"path/filepath"
	"github.com/calavera/dkvolume"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Echo, %q",
		html.EscapeString(r.URL.Path))
}

func volumeContainerCreationHandler () {


}

const testVolumeId = "_test_volume"

var (
	defaultDir  = filepath.Join(dkvolume.DefaultDockerRootDirectory, testVolumeId)	
	root        = flag.String("root", defaultDir, "test volumes root directory")
	wat         = flag.String("wat", "", "wat??")
)


func main() {

	var Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(*wat) == 0 {
		Usage()
		os.Exit(1)
	}

	d := newVolumePoolDriver(*root)
	h := dkvolume.NewHandler(d)
	fmt.Println(h.ServeUnix("root", "test_volume"))
}
