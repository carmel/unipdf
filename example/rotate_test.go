package example

import (
	"fmt"
	"os"
	"testing"

	"github.com/carmel/unipdf/creator"
	"github.com/carmel/unipdf/model"
)

func TestRotate(t *testing.T) {

	var (
		inputPath        = "./assets/2.pdf"
		degrees    int64 = 90
		outputPath       = "rotate.pdf"
	)

	if degrees%90 != 0 {
		fmt.Printf("Degrees needs to be a multiple of 90\n")
		os.Exit(1)
	}

	c := creator.New()

	f, err := os.Open(inputPath)
	checkErr(err)
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	isEncrypted, err := pdfReader.IsEncrypted()
	checkErr(err)

	// Try decrypting both with given password and an empty one if that fails.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		checkErr(err)
		if !auth {
			fmt.Println("Unable to decrypt pdf with empty pass")
			return
		}
	}

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		checkErr(err)

		err = c.AddPage(page)
		checkErr(err)

		_ = c.RotateDeg(degrees)
	}

	c.WriteToFile(outputPath)
}

func TestRotateFlatten(t *testing.T) {
	var (
		inputPath  = "rotate.pdf"
		outputPath = "rotate_flatten.pdf"
	)

	f, err := os.Open(inputPath)
	checkErr(err)
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	isEncrypted, err := pdfReader.IsEncrypted()
	checkErr(err)

	// Try decrypting both with given password and an empty one if that fails.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		checkErr(err)
		if !auth {
			fmt.Println("Unable to decrypt pdf with empty pass")
			return
		}
	}

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	c := creator.New()
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		checkErr(err)

		block, err := creator.NewBlockFromPage(page)
		checkErr(err)

		rotateDeg := int64(0)
		if page.Rotate != nil && *page.Rotate != 0 {
			rotateDeg = 360 - *page.Rotate
		}

		// Rotate the page block if needed.
		if rotateDeg != 0 {
			block.SetAngle(float64(rotateDeg))
		}

		// Set page size in creator.
		// Account for translation that is needed when rotating about the upper left corner.
		if rotateDeg == 90 || rotateDeg == 270 {
			// Swap width and height.
			c.SetPageSize(creator.PageSize{block.Height(), block.Width()})
			block.SetPos(0, block.Width())
		} else {
			c.SetPageSize(creator.PageSize{block.Width(), block.Height()})
			block.SetPos(0, 0)
		}

		c.NewPage()
		err = c.Draw(block)
		checkErr(err)
	}

	c.WriteToFile(outputPath)
}
