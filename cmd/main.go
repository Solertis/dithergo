package main

import (
	"image"
	"image/color"
	"image/png"
	_ "image/jpeg"
	"os"
	"fmt"
	"log"
	"time"
	"github.com/esimov/dithergo"
)

type file struct {
	name string
}

var ditherers []dither.Dither

func (file *file) Open() (image.Image, error) {
	f, err := os.Open(file.name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func (file *file) Grayscale(input image.Image, createImageOutput bool) (*image.Gray, error) {
	bounds := input.Bounds()
	gray := image.NewGray(bounds)

	for x := bounds.Min.X; x < bounds.Dx(); x++ {
		for y := bounds.Min.Y; y < bounds.Dy(); y++ {
			pixel := input.At(x, y)
			gray.Set(x, y, pixel)
		}
	}

	if createImageOutput {
		output, err := os.Create("output/grayscale.png")
		if err != nil {
			return nil, err
		}
		defer output.Close()
		err = png.Encode(output, gray)

		if err != nil {
			log.Fatal(err)
		}
	}

	return gray, nil
}

func (file *file) TresholdDithering(input *image.Gray, createImageOutput bool) (*image.Gray, error) {
	var (
		bounds = input.Bounds()
		dithered = image.NewGray(bounds)
		dx = bounds.Dx()
		dy = bounds.Dy()
	)

	for x := 0; x < dx; x++ {
		for y := 0; y < dy; y++ {
			pixel := input.GrayAt(x, y)
			threshold := func(pixel color.Gray)color.Gray {
				if pixel.Y > 123 {
					return color.Gray{Y:255}
				}
				return color.Gray{Y:0}
			}

			dithered.Set(x, y, threshold(pixel))
		}
	}

	if createImageOutput {
		output, err := os.Create("output/treshold.png")
		if err != nil {
			return nil, err
		}
		defer output.Close()
		err = png.Encode(output, dithered)

		if err != nil {
			log.Fatal(err)
		}
	}

	return dithered, nil
}

func progress(done chan struct{}) {
	ticker := time.NewTicker(time.Millisecond * 100)

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Print(".")
			case <-done:
				ticker.Stop()
			}
		}
	}()
}

func main()  {
	ditherers = []dither.Dither{
		dither.Dither{
			"FloydSteinberg",
			dither.Settings{
				[][]float32{
					[]float32{ 0.0, 0.0, 0.0, 7.0 / 48.0, 5.0 / 48.0 },
					[]float32{ 3.0 / 48.0, 5.0 / 48.0, 7.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0 },
					[]float32{ 1.0 / 48.0, 3.0 / 48.0, 5.0 / 48.0, 3.0 / 48.0, 1.0 / 48.0 },
				},
				0.92,
			},
		},
		dither.Dither{
			"Stucki",
			dither.Settings{
				[][]float32{
					[]float32{ 0.0, 0.0, 0.0, 8.0 / 42.0, 4.0 / 42.0 },
					[]float32{ 2.0 / 42.0, 4.0 / 42.0, 8.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0 },
					[]float32{ 1.0 / 42.0, 2.0 / 42.0, 4.0 / 42.0, 2.0 / 42.0, 1.0 / 42.0 },
				},
				0.92,
			},
		},
		dither.Dither{
			"Athkinson",
			dither.Settings{
				[][]float32{
					[]float32{ 0.0, 0.0, 1.0 / 8.0, 1.0 / 8.0 },
					[]float32{ 1.0 / 8.0, 1.0 / 8.0, 1.0 / 8.0, 0.0 },
					[]float32{ 0.0, 1.0 / 8.0, 0.0, 0.0 },
				},
				0.92,
			},
		},
	}

	if len(os.Args) < 2 || (len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h")) {
		fmt.Println("Usage :  Command --file name", os.Args)
		os.Exit(1)
	}

	done := make(chan struct{})

	input := &file{name: string(os.Args[1])}
	img, _ := input.Open()
	fmt.Print("Rendering image...")

	progress(done)

	func(input *file, done chan struct{}) {
		_ = os.Mkdir("output/color", os.ModePerm)
		_ = os.Mkdir("output/mono", os.ModePerm)
		gray, _ := input.Grayscale(img, true)
		input.TresholdDithering(gray, true)

		for _, ditherer := range ditherers {
			ditherer.PrintColor(img)
			ditherer.PrintMono(img)
		}
		done <-struct{}{}
	}(input, done)

	fmt.Println("\nDone✓")
}