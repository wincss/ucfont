# UCFont: A Go library for reading UCDOS font file

## Usage

    package main

    import (
	    "github.com/wincss/ucfont"
    	"fmt"
	    "os"
    )

    func main() {
	    file, _ := os.Open("UCDOS/UCFONTS/HZKPSST.GBK")
	    f := ucfont.NewPSFontFile(file, true)
	    data, _ := f.GetCharPath('哇')

    	fmt.Println(data)
    }
