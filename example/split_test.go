package example

import (
	"fmt"
	"os"
	"testing"

	"github.com/carmel/unipdf/model"
)

func TestSplit(t *testing.T) {
	var (
		inputPath  = "./assets/2.pdf"
		outputPath = "split.pdf"
		pageFrom   = 2
		pageTo     = 3
	)

	pdfWriter := model.NewPdfWriter()

	f, err := os.Open(inputPath)
	checkErr(err)

	defer f.Close()

	pdfReader, err := model.NewPdfReaderLazy(f)
	checkErr(err)

	isEncrypted, err := pdfReader.IsEncrypted()
	checkErr(err)

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		checkErr(err)
	}

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	if numPages < pageTo {
		fmt.Printf("numPages (%d) < pageTo (%d)\n", numPages, pageTo)
		return
	}

	for i := pageFrom; i <= pageTo; i++ {
		pageNum := i

		page, err := pdfReader.GetPage(pageNum)
		checkErr(err)

		err = pdfWriter.AddPage(page)
		checkErr(err)
	}

	fWrite, err := os.Create(outputPath)
	checkErr(err)

	defer fWrite.Close()

	err = pdfWriter.Write(fWrite)
	checkErr(err)

}

func TestAdvanceSplit(t *testing.T) {
	var (
		inputPath  = "./assets/2.pdf"
		outputPath = "split_advance.pdf"
		pageFrom   = 2
		pageTo     = 3
	)
	pdfWriter := model.NewPdfWriter()

	f, err := os.Open(inputPath)
	checkErr(err)

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	isEncrypted, err := pdfReader.IsEncrypted()
	checkErr(err)

	if isEncrypted {
		_, err = pdfReader.Decrypt([]byte(""))
		checkErr(err)
	}

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	if numPages < pageTo {
		fmt.Printf("numPages (%d) < pageTo (%d)\n", numPages, pageTo)
		return
	}

	// Keep the OC properties intact (optional content).
	// Rarely used but can be relevant in certain cases.
	ocProps, err := pdfReader.GetOCProperties()
	checkErr(err)
	pdfWriter.SetOCProperties(ocProps)

	for i := pageFrom; i <= pageTo; i++ {
		pageNum := i

		page, err := pdfReader.GetPage(pageNum)
		checkErr(err)

		err = pdfWriter.AddPage(page)
		checkErr(err)
	}

	fWrite, err := os.Create(outputPath)
	checkErr(err)

	defer fWrite.Close()

	err = pdfWriter.Write(fWrite)
	checkErr(err)

}
