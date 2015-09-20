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

const testVolumeId = "_pool"

var (
	defaultDir  = filepath.Join(dkvolume.DefaultDockerRootDirectory, testVolumeId)	
	root        = flag.String("root", defaultDir, "the root directory for your volumes")
)


func main() {

	var Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if len(*root) == 0 {
		Usage()
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "root is %s\n", *root)

	d := newVolumePoolDriver(*root)
	h := dkvolume.NewHandler(d)
	fmt.Println(h.ServeUnix("root", "volume_pool"))
}
