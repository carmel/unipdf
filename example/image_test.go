package example

import (
	"archive/zip"
	"fmt"
	"image/jpeg"
	"os"
	"testing"

	"github.com/carmel/unipdf/contentstream"
	"github.com/carmel/unipdf/core"
	"github.com/carmel/unipdf/creator"
	"github.com/carmel/unipdf/extractor"
	"github.com/carmel/unipdf/model"
)

func TestInsertImage(t *testing.T) {
	err := insertImage("./assets/2.pdf", "image_add.pdf", "./assets/1.jpg", 4, 0, 0)
	checkErr(err)
}

func TestAddWatermark(t *testing.T) {
	c := creator.New()

	watermarkImg, err := c.NewImageFromFile("./assets/unidoc-logo.png")
	checkErr(err)

	// Read the input pdf file.
	f, err := os.Open("./assets/2.pdf")
	checkErr(err)
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		// Read the page.
		page, err := pdfReader.GetPage(pageNum)
		checkErr(err)

		// Add to creator.
		c.AddPage(page)

		watermarkImg.ScaleToWidth(c.Context().PageWidth)
		watermarkImg.SetPos(0, (c.Context().PageHeight-watermarkImg.Height())/2)
		watermarkImg.SetOpacity(0.5)
		_ = c.Draw(watermarkImg)
	}

	// Add reader outline tree to the creator.
	c.SetOutlineTree(pdfReader.GetOutlineTree())

	// Add reader AcroForm to the creator.
	c.SetForms(pdfReader.AcroForm)

	c.WriteToFile("watermark.pdf")
}

func TestListImage(t *testing.T) {
	f, err := os.Open("./assets/1.pdf")
	checkErr(err)

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	isEncrypted, err := pdfReader.IsEncrypted()
	checkErr(err)

	if isEncrypted {
		// Try decrypting with an empty one.
		auth, err := pdfReader.Decrypt([]byte(""))
		checkErr(err)
		if !auth {
			fmt.Println("Need to decrypt with a specified user/owner password")
		}
	}

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)
	fmt.Printf("PDF Num Pages: %d\n", numPages)

	for i := 0; i < numPages; i++ {
		fmt.Printf("-----\nPage %d:\n", i+1)

		page, err := pdfReader.GetPage(i + 1)
		checkErr(err)

		// List images on the page.
		err = listImagesOnPage(page)
		checkErr(err)
	}
}

func TestExtractImage(t *testing.T) {
	f, err := os.Open("./assets/1.pdf")
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
			fmt.Println("Need to decrypt with password")
			return
		}
	}

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)
	fmt.Printf("PDF Num Pages: %d\n", numPages)

	// Prepare output archive.
	zipf, err := os.Create("extracted_image.zip")
	checkErr(err)

	defer zipf.Close()
	zipw := zip.NewWriter(zipf)

	totalImages := 0
	for i := 0; i < numPages; i++ {
		fmt.Printf("-----\nPage %d:\n", i+1)

		page, err := pdfReader.GetPage(i + 1)
		checkErr(err)

		pextract, err := extractor.New(page)
		checkErr(err)

		pimages, err := pextract.ExtractPageImages(nil)
		checkErr(err)

		fmt.Printf("%d Images\n", len(pimages.Images))
		for idx, img := range pimages.Images {
			fmt.Printf("Image %d - X: %.2f Y: %.2f, Width: %.2f, Height: %.2f\n",
				totalImages+idx+1, img.X, img.Y, img.Width, img.Height)
			fname := fmt.Sprintf("p%d_%d.jpg", i+1, idx)

			gimg, err := img.Image.ToGoImage()
			checkErr(err)

			imgf, err := zipw.Create(fname)
			checkErr(err)
			opt := jpeg.Options{Quality: 100}
			err = jpeg.Encode(imgf, gimg, &opt)
			checkErr(err)
		}
		totalImages += len(pimages.Images)
	}
	fmt.Printf("Total: %d images\n", totalImages)

	// Make sure to check the error on Close.
	err = zipw.Close()
	checkErr(err)

}

var colorspaces = map[string]int{}
var filters = map[string]int{}

func listImagesOnPage(page *model.PdfPage) error {
	contents, err := page.GetAllContentStreams()
	checkErr(err)

	return listImagesInContentStream(contents, page.Resources)
}

func listImagesInContentStream(contents string, resources *model.PdfPageResources) error {
	cstreamParser := contentstream.NewContentStreamParser(contents)
	operations, err := cstreamParser.Parse()
	checkErr(err)

	processedXObjects := map[string]bool{}

	for _, op := range *operations {
		if op.Operand == "BI" && len(op.Params) == 1 {
			// Inline image.

			iimg, ok := op.Params[0].(*contentstream.ContentStreamInlineImage)
			if !ok {
				continue
			}

			img, err := iimg.ToImage(resources)
			checkErr(err)

			cs, err := iimg.GetColorSpace(resources)
			checkErr(err)

			encoder, err := iimg.GetEncoder()
			checkErr(err)

			fmt.Printf(" Inline image\n")
			fmt.Printf("  Filter: %s\n", encoder.GetFilterName())
			fmt.Printf("  Width: %d\n", img.Width)
			fmt.Printf("  Height: %d\n", img.Height)
			fmt.Printf("  Color components: %d\n", img.ColorComponents)
			fmt.Printf("  ColorSpace: %s\n", cs.String())
			//fmt.Printf("  ColorSpace: %+v\n", cs)
			fmt.Printf("  BPC: %d\n", img.BitsPerComponent)

			// Log filter use globally.
			filter := encoder.GetFilterName()
			filters[filter]++
			// Log colorspace use globally.
			csName := "?"
			if cs != nil {
				csName = cs.String()
			}
			colorspaces[csName]++
		} else if op.Operand == "Do" && len(op.Params) == 1 {
			// XObject.
			name := op.Params[0].(*core.PdfObjectName)

			// Only process each one once.
			_, has := processedXObjects[string(*name)]
			if has {
				continue
			}
			processedXObjects[string(*name)] = true

			_, xtype := resources.GetXObjectByName(*name)
			if xtype == model.XObjectTypeImage {
				fmt.Printf(" XObject Image: %s\n", *name)

				ximg, err := resources.GetXObjectImageByName(*name)
				if err != nil {
					return err
				}
				img, err := ximg.ToImage()
				if err != nil {
					return err
				}

				fmt.Printf("  Filter: %#v\n", ximg.Filter)
				fmt.Printf("  Width: %v\n", *ximg.Width)
				fmt.Printf("  Height: %d\n", *ximg.Height)
				fmt.Printf("  Color components: %d\n", img.ColorComponents)
				fmt.Printf("  ColorSpace: %s\n", ximg.ColorSpace.String())
				fmt.Printf("  ColorSpace: %#v\n", ximg.ColorSpace)
				fmt.Printf("  BPC: %v\n", *ximg.BitsPerComponent)

				// Log filter use globally.
				filter := ximg.Filter.GetFilterName()
				filters[filter]++
				// Log colorspace use globally.
				cs := ximg.ColorSpace.String()
				colorspaces[cs]++
			} else if xtype == model.XObjectTypeForm {
				// Go through the XObject Form content stream.
				fmt.Printf("--> XObject Form: %s\n", *name)
				xform, err := resources.GetXObjectFormByName(*name)
				if err != nil {
					return err
				}

				formContent, err := xform.GetContentStream()
				if err != nil {
					return err
				}
				fmt.Printf("xform: %#v\n", xform)
				fmt.Printf("xform res: %#v\n", xform.Resources)
				fmt.Printf("Content: %s\n", formContent)

				// Process the content stream in the Form object too:
				// XXX/TODO: Use either form resources (priority) and fall back to page resources alternatively if not found.
				if xform.Resources != nil {
					err = listImagesInContentStream(string(formContent), xform.Resources)
				} else {
					err = listImagesInContentStream(string(formContent), resources)
				}
				if err != nil {
					return err
				}
				fmt.Printf("<-- XObject Form: %s\n", *name)
			}
		}
	}

	return nil
}

// Add image to a specific page of a PDF.  xPos and yPos define the upper left corner of the image location, and iwidth
// is the width of the image in PDF document dimensions (height/width ratio is maintained).
func insertImage(inputPath string, outputPath string, imagePath string, pageNum int, xPos float64, yPos float64) error {

	c := creator.New()

	// Prepare the image.
	img, err := c.NewImageFromFile(imagePath)
	checkErr(err)
	img.ScaleToWidth(img.Width())
	img.SetPos(xPos, yPos)

	// Read the input pdf file.
	f, err := os.Open(inputPath)
	checkErr(err)
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	checkErr(err)

	numPages, err := pdfReader.GetNumPages()
	checkErr(err)

	addPage := func(init, num int) error {
		for i := init; i <= num; i++ {
			page, err := pdfReader.GetPage(i)
			checkErr(err)

			// Add the page.
			err = c.AddPage(page)
			checkErr(err)
		}
		return nil
	}

	switch true {
	case pageNum == 1:
		c.Draw(img)
		addPage(1, numPages)
	case pageNum == numPages+1:
		addPage(1, numPages)
		c.NewPage()
		c.Draw(img)
	case pageNum > 1 && pageNum <= numPages:
		addPage(1, pageNum-1)
		c.NewPage()
		c.Draw(img)
		addPage(pageNum, numPages)
	}

	err = c.WriteToFile(outputPath)
	return err
}
