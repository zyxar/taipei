package main

import (
	"flag"
	"fmt"
	"github.com/zyxar/taipei"
	// "io/ioutil"
)

func verify_bt(filename, path string) error {
	m, err := taipei.GetMetaInfo(filename)
	if err != nil {
		return err
	}
	fmt.Println(m)
	taipei.SetEcho(true)
	g, err := taipei.VerifyContent(m, path)
	taipei.SetEcho(false)
	if g == false {
		return err
	}
	return nil
}

func main() {
	var torfile, pathd string
	flag.StringVar(&torfile, "tor", "", "Torrent File")
	flag.StringVar(&pathd, "path", "", "Bt content directory")
	flag.Parse()
	if len(torfile) == 0 || len(pathd) == 0 {
		fmt.Println("Invalid torrent file or bt content directory")
		flag.Usage()
		return
	}
	err := verify_bt(torfile, pathd)
	if err != nil {
		fmt.Println(err)
	}
}
