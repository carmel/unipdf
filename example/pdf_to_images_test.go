package example_test

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carmel/unipdf/model"
	"github.com/carmel/unipdf/render"
)

func TestPdf2Images(t *testing.T) {

	filename := "unidoc-report.pdf"
	// Create reader.
	reader, f, err := model.NewPdfReaderFromFile(filename)
	if err != nil {
		log.Fatalf("Could not create reader: %v\n", err)
	}
	defer reader.Close()

	// Get total number of pages.
	numPages, err := f.GetNumPages()
	if err != nil {
		log.Fatalf("Could not retrieve number of pages: %v\n", err)
	}

	// Render pages.
	basename := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))

	device := render.NewImageDevice()
	for i := 1; i <= numPages; i++ {
		// Get page.
		page, err := f.GetPage(i)
		if err != nil {
			log.Fatalf("Could not retrieve page: %v\n", err)
		}

		// Render page to PNG file.
		// RenderToPath chooses the image format by looking at the extension
		// of the specified filename. Only PNG and JPEG files are supported
		// currently.
		outFilename := filepath.Join("output", fmt.Sprintf("%s_%d.png", basename, i))
		if err = device.RenderToPath(page, outFilename); err != nil {
			log.Fatalf("Image rendering error: %v\n", err)
		}

		// Alternatively, an image.Image instance can be obtained by using
		// the Render method of the image device, which can then be encoded
		// and saved in any format.
		// image, err := device.Render(page)
	}
}
