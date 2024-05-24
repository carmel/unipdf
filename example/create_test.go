package example

import (
	"bytes"
	"fmt"
	"image"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/carmel/unipdf/common"
	"github.com/carmel/unipdf/creator"
	"github.com/carmel/unipdf/model"
	"github.com/wcharczuk/go-chart"
)

func TestCreatePDFReport(t *testing.T) {
	robotoFontRegular, err := model.NewPdfFontFromTTFFile("./assets/Roboto-Regular.ttf")
	checkErr(err)

	robotoFontPro, err := model.NewPdfFontFromTTFFile("./assets/Roboto-Bold.ttf")
	checkErr(err)

	c := creator.New()
	c.SetPageMargins(50, 50, 100, 70)
	c.SetPageSize(creator.PageSizeA4)

	// Generate the table of contents.
	c.AddTOC = true
	toc := c.TOC()
	hstyle := c.NewTextStyle()
	hstyle.Color = creator.ColorRGBFromArithmetic(0.2, 0.2, 0.2)
	hstyle.FontSize = 28
	toc.SetHeading("Table of Contents", hstyle)
	lstyle := c.NewTextStyle()
	lstyle.FontSize = 14
	toc.SetLineStyle(lstyle)

	logoImg, err := c.NewImageFromFile("./assets/unidoc-logo.png")
	checkErr(err)

	logoImg.ScaleToHeight(25)
	logoImg.SetPos(58, 20)

	documentControlPage(c, robotoFontRegular, robotoFontPro)

	featureOverviewPage(c, robotoFontRegular, robotoFontPro)

	// Setup a front page (always placed first).
	c.CreateFrontPage(func(args creator.FrontpageFunctionArgs) {
		frontPage(c)
	})

	// Draw a header on each page.
	c.DrawHeader(func(block *creator.Block, args creator.HeaderFunctionArgs) {
		// Draw the header on a block. The block size is the size of the page's top margins.
		block.Draw(logoImg)
	})

	// Draw footer on each page.
	c.DrawFooter(func(block *creator.Block, args creator.FooterFunctionArgs) {
		// Draw the on a block for each page.
		p := c.NewParagraph("unidoc.io")
		p.SetFont(robotoFontRegular)
		p.SetFontSize(8)
		p.SetPos(50, 20)
		p.SetColor(creator.ColorRGBFrom8bit(63, 68, 76))
		block.Draw(p)

		strPage := fmt.Sprintf("Page %d of %d", args.PageNum, args.TotalPages)
		p = c.NewParagraph(strPage)
		p.SetFont(robotoFontRegular)
		p.SetFontSize(8)
		p.SetPos(300, 20)
		p.SetColor(creator.ColorRGBFrom8bit(63, 68, 76))
		block.Draw(p)
	})

	err = c.WriteToFile("unidoc-report.pdf")
	checkErr(err)

}

func TestCreateFromImages(t *testing.T) {

	images := []string{"./assets/1.jpg", "./assets/2.jpg"}
	c := creator.New()

	for _, imgPath := range images {

		img, err := c.NewImageFromFile(imgPath)
		checkErr(err)
		img.ScaleToWidth(612.0)

		// Use page width of 612 points, and calculate the height proportionally based on the image.
		// Standard PPI is 72 points per inch, thus a width of 8.5"
		height := 612.0 * img.Height() / img.Width()
		c.SetPageSize(creator.PageSize{612, height})
		c.NewPage()
		img.SetPos(0, 0)
		_ = c.Draw(img)
	}

	err := c.WriteToFile("from_images.pdf")
	checkErr(err)
}

// Generates the front page.
func frontPage(c *creator.Creator) {
	helvetica, _ := model.NewStandard14Font("Helvetica")
	helveticaBold, _ := model.NewStandard14Font("Helvetica-Bold")

	p := c.NewParagraph("UniDoc")
	p.SetFont(helvetica)
	p.SetFontSize(48)
	p.SetMargins(85, 0, 150, 0)
	p.SetColor(creator.ColorRGBFrom8bit(56, 68, 77))
	c.Draw(p)

	p = c.NewParagraph("Example Report")
	p.SetFont(helveticaBold)
	p.SetFontSize(30)
	p.SetMargins(85, 0, 0, 0)
	p.SetColor(creator.ColorRGBFrom8bit(45, 148, 215))
	c.Draw(p)

	t := time.Now().UTC()
	dateStr := t.Format("1 Jan, 2006 15:04")

	p = c.NewParagraph(dateStr)
	p.SetFont(helveticaBold)
	p.SetFontSize(12)
	p.SetMargins(90, 0, 5, 0)
	p.SetColor(creator.ColorRGBFrom8bit(56, 68, 77))
	c.Draw(p)
}

// Document control page.
func documentControlPage(c *creator.Creator, fontRegular *model.PdfFont, fontBold *model.PdfFont) {
	ch := c.NewChapter("Document control")
	ch.SetMargins(0, 0, 40, 0)
	ch.GetHeading().SetFont(fontRegular)
	ch.GetHeading().SetFontSize(18)
	ch.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))

	sc := ch.NewSubchapter("Issuer details")
	sc.GetHeading().SetFont(fontRegular)
	sc.GetHeading().SetFontSize(18)
	sc.GetHeading().SetColor(creator.ColorRGBFrom8bit(72, 86, 95))

	issuerTable := c.NewTable(2)
	issuerTable.SetMargins(0, 0, 30, 0)

	pColor := creator.ColorRGBFrom8bit(72, 86, 95)
	bgColor := creator.ColorRGBFrom8bit(56, 68, 67)

	p := c.NewParagraph("Issuer")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell := issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = c.NewParagraph("UniDoc")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetContent(p)

	p = c.NewParagraph("Address")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = c.NewParagraph("Klapparstig 16, 101 Reykjavik, Iceland")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetContent(p)

	p = c.NewParagraph("Email")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBackgroundColor(bgColor)
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetContent(p)

	p = c.NewParagraph("sales@unidoc.io")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetContent(p)

	p = c.NewParagraph("Web")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = c.NewParagraph("unidoc.io")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetContent(p)

	p = c.NewParagraph("Author")
	p.SetFont(fontBold)
	p.SetFontSize(10)
	p.SetColor(creator.ColorWhite)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetBackgroundColor(bgColor)
	cell.SetContent(p)

	p = c.NewParagraph("UniDoc report generator")
	p.SetFont(fontRegular)
	p.SetFontSize(10)
	p.SetColor(pColor)
	cell = issuerTable.NewCell()
	cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
	cell.SetContent(p)

	sc.Add(issuerTable)

	// 1.2 - Document history
	sc = ch.NewSubchapter("Document History")
	sc.SetMargins(0, 0, 5, 0)
	sc.GetHeading().SetFont(fontRegular)
	sc.GetHeading().SetFontSize(18)
	sc.GetHeading().SetColor(pColor)

	histTable := c.NewTable(3)
	histTable.SetMargins(0, 0, 30, 50)

	histCols := []string{"Date Issued", "UniDoc Version", "Type/Change"}
	for _, histCol := range histCols {
		p = c.NewParagraph(histCol)
		p.SetFont(fontBold)
		p.SetFontSize(10)
		p.SetColor(creator.ColorWhite)
		cell = histTable.NewCell()
		cell.SetBackgroundColor(bgColor)
		cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
		cell.SetHorizontalAlignment(creator.CellHorizontalAlignmentCenter)
		cell.SetVerticalAlignment(creator.CellVerticalAlignmentMiddle)
		cell.SetContent(p)
	}

	dateStr := common.ReleasedAt.Format("1 Jan, 2006 15:04")

	histVals := []string{dateStr, common.Version, "First issue"}
	for _, histVal := range histVals {
		p = c.NewParagraph(histVal)
		p.SetFont(fontRegular)
		p.SetFontSize(10)
		p.SetColor(pColor)
		cell = histTable.NewCell()
		cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
		cell.SetHorizontalAlignment(creator.CellHorizontalAlignmentCenter)
		cell.SetVerticalAlignment(creator.CellVerticalAlignmentMiddle)
		cell.SetContent(p)
	}

	sc.Add(histTable)

	err := c.Draw(ch)
	if err != nil {
		panic(err)
	}
}

// Chapter giving an overview of features.
// TODO: Add code snippets and show more styles and options.
func featureOverviewPage(c *creator.Creator, fontRegular *model.PdfFont, fontBold *model.PdfFont) {
	// Ensure that the chapter starts on a new page.
	c.NewPage()

	ch := c.NewChapter("Feature overview")

	chapterFont := fontRegular
	chapterFontColor := creator.ColorRGBFrom8bit(72, 86, 95)
	chapterFontSize := 18.0

	normalFont := fontRegular
	normalFontColor := creator.ColorRGBFrom8bit(72, 86, 95)
	normalFontSize := 10.0

	bgColor := creator.ColorRGBFrom8bit(56, 68, 67)

	ch.GetHeading().SetFont(chapterFont)
	ch.GetHeading().SetFontSize(chapterFontSize)
	ch.GetHeading().SetColor(chapterFontColor)

	p := c.NewParagraph("This chapter demonstrates a few of the features of UniDoc that can be used for report generation.")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	ch.Add(p)

	// Paragraphs.
	sc := ch.NewSubchapter("Paragraphs")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = c.NewParagraph("Paragraphs are used to represent text, as little as a single character, a word or " +
		"multiple words forming multiple sentences. UniDoc handles automatically wrapping those across lines and pages, making " +
		"it relatively easy to work with. They can also be left, center, right aligned or justified as illustrated below:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	// Example paragraphs:
	loremTxt := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt" +
		"ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut " +
		"aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore" +
		"eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt " +
		"mollit anim id est laborum."
	alignments := []creator.TextAlignment{creator.TextAlignmentLeft, creator.TextAlignmentCenter,
		creator.TextAlignmentRight, creator.TextAlignmentJustify}
	for j := 0; j < 4; j++ {
		p = c.NewParagraph(loremTxt)
		p.SetFont(normalFont)
		p.SetFontSize(normalFontSize)
		p.SetColor(normalFontColor)
		p.SetMargins(20, 0, 10, 10)
		p.SetTextAlignment(alignments[j%4])

		sc.Add(p)
	}

	sc = ch.NewSubchapter("Tables")
	// Mock table: Priority table.
	priTable := c.NewTable(2)
	priTable.SetMargins(40, 40, 10, 0)
	// Column headers:
	tableCols := []string{"Priority", "Items fulfilled / available"}
	for _, tableCol := range tableCols {
		p = c.NewParagraph(tableCol)
		p.SetFont(fontBold)
		p.SetFontSize(10)
		p.SetColor(creator.ColorWhite)
		cell := priTable.NewCell()
		cell.SetBackgroundColor(bgColor)
		cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
		cell.SetContent(p)
	}
	items := [][]string{
		{"High", "52/80"},
		{"Medium", "32/100"},
		{"Low", "10/90"},
	}
	for _, lineItems := range items {
		for _, item := range lineItems {
			p = c.NewParagraph(item)
			p.SetFont(fontBold)
			p.SetFontSize(10)
			p.SetColor(creator.ColorWhite)
			cell := priTable.NewCell()
			cell.SetBackgroundColor(bgColor)
			cell.SetBorder(creator.CellBorderSideAll, creator.CellBorderStyleSingle, 1)
			cell.SetContent(p)
		}
	}
	sc.Add(priTable)

	sc = ch.NewSubchapter("Images")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = c.NewParagraph("Images can be loaded from multiple file formats, example from a PNG image:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 5)
	sc.Add(p)

	// Show logo.
	img, err := c.NewImageFromFile("./assets/unidoc-logo.png")
	if err != nil {
		panic(err)
	}
	img.ScaleToHeight(50)
	sc.Add(img)

	sc = ch.NewSubchapter("QR Codes / Barcodes")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = c.NewParagraph("Example of a QR code generated with package github.com/boombuler/barcode:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 5)
	sc.Add(p)

	qrCode, _ := makeQrCodeImage("HELLO", 40, 5)
	img, err = c.NewImageFromGoImage(qrCode)
	if err != nil {
		panic(err)
	}
	img.SetWidth(40)
	img.SetHeight(40)
	sc.Add(img)

	sc = ch.NewSubchapter("Graphing / Charts")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = c.NewParagraph("Graphs can be generated via packages such as github.com/wcharczuk/go-chart as illustrated " +
		"in the following plot:")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	graph := chart.PieChart{
		Width:  200,
		Height: 200,
		Values: []chart.Value{
			{Value: 70, Label: "Compliant"},
			{Value: 30, Label: "Non-Compliant"},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err = graph.Render(chart.PNG, buffer)
	if err != nil {
		panic(err)
	}
	img, err = c.NewImageFromData(buffer.Bytes())
	if err != nil {
		panic(err)
	}
	img.SetMargins(0, 0, 10, 0)
	sc.Add(img)

	sc = ch.NewSubchapter("Headers and footers")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = c.NewParagraph("Convenience functions are provided to generate headers and footers, see: " +
		"https://godoc.org/github.com/unidoc/unipdf/creator#Creator.DrawHeader and " +
		"https://godoc.org/github.com/unidoc/unipdf/creator#Creator.DrawFooter " +
		"They both set a function that accepts a block which the header/footer is drawn on for each page. " +
		"More information is provided in the arguments, allowing to skip header/footer on specific pages and " +
		"showing page number and count.")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	sc = ch.NewSubchapter("Table of contents generation")
	sc.GetHeading().SetMargins(0, 0, 20, 0)
	sc.GetHeading().SetFont(chapterFont)
	sc.GetHeading().SetFontSize(chapterFontSize)
	sc.GetHeading().SetColor(chapterFontColor)

	p = c.NewParagraph("A convenience function is provided to generate table of contents " +
		"as can be seen on https://godoc.org/github.com/unidoc/unipdf/creator#Creator.CreateTableOfContents and " +
		"in our example code on unidoc.io.")
	p.SetFont(normalFont)
	p.SetFontSize(normalFontSize)
	p.SetColor(normalFontColor)
	p.SetMargins(0, 0, 5, 0)
	sc.Add(p)

	c.Draw(ch)
}

// Helper function to make the QR code image with a specified oversampling factor.
// The oversampling specifies how many pixels/point. Standard PDF resolution is 72 points/inch.
func makeQrCodeImage(text string, width float64, oversampling int) (image.Image, error) {
	qrCode, err := qr.Encode(text, qr.M, qr.Auto)
	if err != nil {
		return nil, err
	}

	pixelWidth := oversampling * int(math.Ceil(width))
	qrCode, err = barcode.Scale(qrCode, pixelWidth, pixelWidth)
	if err != nil {
		return nil, err
	}

	return qrCode, nil
}

func TestCreateInvoices(t *testing.T) {
	// Instantiate new PDF creator
	c := creator.New()

	// Create a new PDF page and select it for editing
	c.NewPage()

	// Create new invoice and populate it with data
	invoice := createInvoice(c, "./assets/unidoc-logo.png")

	// Write invoice to page
	checkErr(c.Draw(invoice))

	// Write output file.
	// Alternative is writing to a Writer interface by using c.Write
	checkErr(c.WriteToFile("nvoice.pdf"))
}

func createInvoice(c *creator.Creator, logoPath string) *creator.Invoice {
	// Create an instance of Logo used as a header for the invoice
	// If the image is not stored localy, you can use NewImageFromData to generate it from byte array
	logo, err := c.NewImageFromFile(logoPath)
	checkErr(err)

	// Create a new invoice
	invoice := c.NewInvoice()

	// Set invoice logo
	invoice.SetLogo(logo)

	// Set invoice information
	invoice.SetNumber("0001")
	invoice.SetDate("28/07/2016")
	invoice.SetDueDate("28/07/2016")
	invoice.AddInfo("Payment terms", "Due on receipt")
	invoice.AddInfo("Paid", "No")

	// Set invoice addresses
	invoice.SetSellerAddress(&creator.InvoiceAddress{
		Name:    "John Doe",
		Street:  "8 Elm Street",
		City:    "Cambridge",
		Zip:     "CB14DH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "johndoe@email.com",
	})

	invoice.SetBuyerAddress(&creator.InvoiceAddress{
		Name:    "Jane Doe",
		Street:  "9 Elm Street",
		City:    "London",
		Zip:     "LB15FH",
		Country: "United Kingdom",
		Phone:   "xxx-xxx-xxxx",
		Email:   "janedoe@email.com",
	})

	// Add products to invoice
	for i := 1; i < 6; i++ {
		invoice.AddLine(
			fmt.Sprintf("Test product #%d", i),
			"1",
			strconv.Itoa((i-1)*7),
			strconv.Itoa((i+4)*3),
		)
	}

	// Set invoice totals
	invoice.SetSubtotal("$100.00")
	invoice.AddTotalLine("Tax (10%)", "$10.00")
	invoice.AddTotalLine("Shipping", "$5.00")
	invoice.SetTotal("$115.00")

	// Set invoice content sections
	invoice.SetNotes("Notes", "Thank you for your business.")
	invoice.SetTerms("Terms and conditions", "Full refund for 60 days after purchase.")

	return invoice
}
