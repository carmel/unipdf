package example

import (
	"fmt"
	"os"
	"testing"

	"github.com/carmel/unipdf/extractor"
	"github.com/carmel/unipdf/model"
)

func TestPageProperties(t *testing.T) {

	var (
		inputPath = "./assets/2.pdf"
		pageNum   = 2
	)

	f, err := os.Open(inputPath)
	checkErr(err)
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	isEncrypted, err := pdfReader.IsEncrypted()
	checkErr(err)

	// Try decrypting with an empty one.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		checkErr(err)
		if !auth {
			fmt.Println("Encrypted - unable to access - update code to specify pass")
			return
		}
	}

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	// If invalid pagenum, print all pages.
	if pageNum <= 0 || pageNum > numPages {
		for i := 0; i < numPages; i++ {
			page, err := pdfReader.GetPage(i + 1)
			checkErr(err)
			fmt.Printf("-- Page %d\n", i+1)
			err = processPage(page)
			checkErr(err)
		}
	} else {
		page, err := pdfReader.GetPage(pageNum)
		checkErr(err)
		fmt.Printf("-- Page %d\n", pageNum)
		err = processPage(page)
		checkErr(err)
	}

}

func TestPageText(t *testing.T) {
	var (
		inputPath = "./assets/2.pdf"
	)

	f, err := os.Open(inputPath)
	checkErr(err)

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	fmt.Printf("--------------------\n")
	fmt.Printf("PDF to text extraction:\n")
	fmt.Printf("--------------------\n")
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		checkErr(err)

		ex, err := extractor.New(page)
		checkErr(err)

		text, err := ex.ExtractText()
		checkErr(err)

		fmt.Println("------------------------------")
		fmt.Printf("Page %d:\n", pageNum)
		fmt.Printf("\"%s\"\n", text)
		fmt.Println("------------------------------")
	}
}

func processPage(page *model.PdfPage) error {
	mBox, err := page.GetMediaBox()
	checkErr(err)
	pageWidth := mBox.Urx - mBox.Llx
	pageHeight := mBox.Ury - mBox.Lly

	fmt.Printf(" Page: %+v\n", page)
	if page.Rotate != nil {
		fmt.Printf(" Page rotation: %v\n", *page.Rotate)
	} else {
		fmt.Printf(" Page rotation: 0\n")
	}
	fmt.Printf(" Page mediabox: %+v\n", page.MediaBox)
	fmt.Printf(" Page height: %f\n", pageHeight)
	fmt.Printf(" Page width: %f\n", pageWidth)

	return nil
}
