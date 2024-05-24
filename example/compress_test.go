package example

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/carmel/unipdf/model"
	"github.com/carmel/unipdf/model/optimize"
)

func TestCompress(t *testing.T) {

	var (
		inputPath  = "./assets/2.pdf"
		outputPath = "compress.pdf"
	)

	// Initialize starting time.
	start := time.Now()

	// Get input file stat.
	inputFileInfo, err := os.Stat(inputPath)
	checkErr(err)

	// Create reader.
	inputFile, err := os.Open(inputPath)
	checkErr(err)
	defer inputFile.Close()

	reader, err := model.NewPdfReader(inputFile)
	checkErr(err)

	// Get number of pages in the input file.
	pages, err := reader.GetNumPages()
	checkErr(err)

	// Add input file pages to the writer.
	writer := model.NewPdfWriter()
	for i := 1; i <= pages; i++ {
		page, err := reader.GetPage(i)
		checkErr(err)

		err = writer.AddPage(page)
		checkErr(err)
	}

	// Add reader AcroForm to the writer.
	if reader.AcroForm != nil {
		writer.SetForms(reader.AcroForm)
	}

	// Set optimizer.
	writer.SetOptimizer(optimize.New(optimize.Options{
		CombineDuplicateDirectObjects:   true,
		CombineIdenticalIndirectObjects: true,
		CombineDuplicateStreams:         true,
		CompressStreams:                 true,
		UseObjectStreams:                true,
		ImageQuality:                    80,
		ImageUpperPPI:                   100,
	}))

	// Create output file.
	outputFile, err := os.Create(outputPath)
	checkErr(err)
	defer outputFile.Close()

	// Write output file.
	err = writer.Write(outputFile)
	checkErr(err)

	// Get output file stat.
	outputFileInfo, err := os.Stat(outputPath)
	checkErr(err)

	// Print basic optimization statistics.
	inputSize := inputFileInfo.Size()
	outputSize := outputFileInfo.Size()
	ratio := 100.0 - (float64(outputSize) / float64(inputSize) * 100.0)
	duration := float64(time.Since(start)) / float64(time.Millisecond)

	fmt.Printf("Original file: %s\n", inputPath)
	fmt.Printf("Original size: %d bytes\n", inputSize)
	fmt.Printf("Optimized file: %s\n", outputPath)
	fmt.Printf("Optimized size: %d bytes\n", outputSize)
	fmt.Printf("Compression ratio: %.2f%%\n", ratio)
	fmt.Printf("Processing time: %.2f ms\n", duration)
}
