package main

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"path"

	"github.com/unidoc/unipdf/extractor"
	"github.com/unidoc/unipdf/model"
)

// Config - Info from config file
type Config struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// Reads info from config file
func readConfig(jsonFilePath string) (config Config) {

	// read our opened xmlFile as a byte array.
	byteValue, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		fmt.Println(err)
	}

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	_ = json.Unmarshal(byteValue, &config)
	return
}

func main() {
	// Enable debug-level console logging, when debuggingn:
	//unicommon.SetLogger(unicommon.NewConsoleLogger(unicommon.LogLevelDebug))

	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	config := readConfig(path.Join(mydir, "/resources/config.json"))
	fmt.Printf("Input file: %s\n", config.Input)
	err = extractImagesToArchive(config.Input, config.Output)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// Extracts images and properties of a PDF specified by inputPath.
// The output images are stored into a zip archive whose path is given by outputPath.
func extractImagesToArchive(inputPath, outputPath string) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return err
	}

	isEncrypted, err := pdfReader.IsEncrypted()
	if err != nil {
		return err
	}

	// Try decrypting with an empty one.
	if isEncrypted {
		auth, err := pdfReader.Decrypt([]byte(""))
		if err != nil {
			// Encrypted and we cannot do anything about it.
			return err
		}
		if !auth {
			fmt.Println("Need to decrypt with password")
			return nil
		}
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return err
	}
	fmt.Printf("PDF Num Pages: %d\n", numPages)

	totalImages := 0
	for i := 0; i < numPages; i++ {
		fmt.Printf("-----\nPage %d:\n", i+1)

		page, err := pdfReader.GetPage(i + 1)
		if err != nil {
			return err
		}

		pextract, err := extractor.New(page)
		if err != nil {
			return err
		}

		pimages, err := pextract.ExtractPageImages(nil)
		if err != nil {
			return err
		}

		fmt.Printf("%d Images\n", len(pimages.Images))
		for idx, img := range pimages.Images {
			fname := fmt.Sprintf("page_%d_%d.png", i+1, idx)

			gimg, err := img.Image.ToGoImage()
			if err != nil {
				return err
			}

			imgf, _ := os.Create(path.Join(outputPath, fname))
			err = png.Encode(imgf, gimg)
			if err != nil {
				return err
			}
		}
		totalImages += len(pimages.Images)
	}
	fmt.Printf("Total: %d images\n", totalImages)

	return nil
}
