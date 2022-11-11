package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/color"
	"os"
	"strconv"
	"strings"
	"sync"
)


// DEFINE COLOR PALETTES

// Black and White
var oneBit = color.Palette([]color.Color{
	color.RGBA{0, 0, 0, 255},
	color.RGBA{255, 255, 255, 255}})
// Greyscale
var greys = color.Palette([]color.Color{
	color.RGBA{51, 51, 51, 255},
	color.RGBA{102, 102, 102, 255},
	color.RGBA{153, 153, 153, 255},
	color.RGBA{204, 204, 204, 255}})
// Green-Grey Gradient
var gameboy = color.Palette([]color.Color{
	color.RGBA{8, 24, 32, 255},
	color.RGBA{52, 104, 86, 255},
	color.RGBA{136, 192, 112, 255},
	color.RGBA{224, 248, 208, 255}})
// Black and Lime Green
var retro = color.Palette([]color.Color{
	color.RGBA{40, 40, 40, 255},
	color.RGBA{51, 255, 51, 255}})
// Blue and White
var aqua = color.Palette([]color.Color{
	color.RGBA{0, 128, 191, 255},
	color.RGBA{0, 172, 223, 255},
	color.RGBA{85, 208, 255, 255},
	color.RGBA{124, 232, 255, 255}})
// Reds and Browns
var warm = color.Palette([]color.Color{
	color.RGBA{100, 69, 54, 255},
	color.RGBA{178, 103, 94, 255},
	color.RGBA{196, 163, 129, 255},
	color.RGBA{238, 241, 189, 255}})


// Given an RGBA color, return an 8-bit grayscale equivalent
func findGray(inPix color.Color) uint8 {
	// Get RGB values from pixel (uint32)
	r, g, b, _ := inPix.RGBA()

	/*
	NTSC Formula—RGB to Grayscale
	Gray = (0.299 * Red) + (0.587 * Green) + (0.114 * Blue)

	Different colors have different percieved brightness levels
	These weights match the perception of an average person
	*/

	gray := (0.299 * float64(r>>8)) + (0.587 * float64(g>>8)) + (0.114 * float64(b>>8))

	return uint8(gray)
}


/*
grayVal = 8-bit grayscale value at a neighboring pixel
quantError = oldColor - newColor (for current pixel)
numerator = Portion of error distributed, divided by 16
*/
func pushError(grayVal uint8, quantError int, numerator int) uint8 {

	// Add error to neighboring pixel
	var newPix float32 = float32(grayVal) + (float32(quantError) * (float32(numerator)/16))

	// Clip values into 0-255 range
	if newPix > 255 {
		return 255
	} else if newPix < 0 {
		return 0
	}

	return uint8(newPix)
}


func dither(i int, frame *image.Paletted, palette color.Palette, paletteSize int, imgArr []*image.Paletted, wg *sync.WaitGroup) {
	// Signal to WaitGroup after function returns
	defer wg.Done()

	/* 
	Create variables for image size
	*/
	bounds := frame.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	/*
	Create a grayscale representation of the pixel grid
	Uses a 2D slice which allows for non-constant size declarations
	*/
	grayPix := make([][]uint8, h)						// Declare 2D slice
	for y:=range grayPix {								// For each y-value
	    grayPix[i] = make([]uint8, w)					// Create slice for row
	    for x:=0; x<w; x++ {							// For each x-value
	    	grayPix[y][x] = findGray(frame.At(x, y))	// Convert to greyscale
	    }
	}

	/*
	Create a new paletted image which will hold the output frame
	Since the image type is Paletted, the RGBA colors to be used are already defined
	*/
	newImg := image.NewPaletted(frame.Bounds(), palette)

	/*
	Iterate over each pixel in the image
	https://en.wikipedia.org/wiki/Floyd%E2%80%93Steinberg_dithering
	*/
	for y:=0; y<h; y++ {
		for x:=0; x<w; x++ {

			/*
			First, determine the current pixel value (grayscale)
			and declare a variable to hold the new pixel value
			*/
			oldPix := grayPix[y][x]
			var newPix uint8

			/*
			Set value for new pixel
			Based off the amount of colors in the palette
			*/
			if paletteSize == 4 {
				if oldPix < 64 {
					newPix = 31
					newImg.Set(x, y, palette[0])
				} else if oldPix < 128 {
					newPix = 95
					newImg.Set(x, y, palette[1])
				} else if oldPix < 192 {
					newPix = 159
					newImg.Set(x, y, palette[2])
				} else {
					newPix = 223
					newImg.Set(x, y, palette[3])
				}
			} else {
				if oldPix < 128 {
					newPix = 31
					newImg.Set(x, y, palette[0])
				} else {
					newPix = 223
					newImg.Set(x, y, palette[1])
				}
			}

			/*
			Calculate error between original and new colors
			*/
			var quantError int = int(oldPix)-int(newPix)

			// Floyd-Steinberg "pushes" the error from each pixel to the surrounding pixels
			/*
				[0] [*] [7]
				[3] [5] [1]
				The pixel being worked on is represented with '*'
				Portions of the error (the cell number/16) is added
				to 4 of the neighboring cells
			*/
			// Only push error to existing pixels
			// Have to check that we aren't at the right edge or bottom row
			if x+1 < w {
				grayPix[y][x+1] = pushError(grayPix[y][x+1], quantError, 7)
			}
			if y+1 < h {
				if x-1 > 0 {
					grayPix[y+1][x-1] = pushError(grayPix[y+1][x-1], quantError, 3)
				}
				grayPix[y+1][x] = pushError(grayPix[y+1][x], quantError, 5)
				if x+1 < w {
					grayPix[y+1][x+1] = pushError(grayPix[y+1][x+1], quantError, 1)
				}
			}
		}
	}

	/*
	To avoid having to sync the routine returns—
	Dereference pointer to output array, add frame to correct index
	*/
	imgArr[i] = newImg
	return
}


func main() {
	/*
	Check provided arguments
	Ensure that there is both an input and output
	Both input and output must be .gif files

	Proper usage—
	go run ditherGif.go [infile] [outfile] [palette (1-6)]

	Palette argument is optional, defaults to 1
	*/
	if len(os.Args) < 3 || !strings.HasSuffix(os.Args[1], ".gif") || !strings.HasSuffix(os.Args[2], ".gif") {
		fmt.Println("Usage:\ngo run dither.go [infile] [outfile] [palette (1-6)]")
		os.Exit(1)
	}

	/*
	Determine palette
	Defaults to 1-bit black and white
	Values 2-6 correspond to other palettes
	*/
	var palette = oneBit
	if len(os.Args) == 4 {								// If palette argument is present
		paletteArg, err := strconv.Atoi(os.Args[3])		// Convert argument to an int
		if err != nil {
			fmt.Println("Invalid palette argument (expecting value 1-6)")
			os.Exit(1)
		}
		switch paletteArg {								// Set proper palette
		case 2: 
			palette = greys
		case 3:
			palette = gameboy
		case 4:
			palette = retro
		case 5:
			palette = aqua
		case 6:
			palette = warm
		}
	}
	var paletteSize int = len(palette)

	/*
	Open image file
	*/
	img, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Could not open", os.Args[1])
		os.Exit(1)
	}
	defer img.Close()

	/*
	Decode image data
	.gif files are a series of 256-color images
	Each image is represented by an int array
	*/
	imgData, _ := gif.DecodeAll(img)

	/*
	Define struct to hold output .gif data
	Image—Pointer to array of images
	Delay—Integer defining delay time between frames
	*/
	imgArr := make([]*image.Paletted, len(imgData.Image))
	anim := &gif.GIF{Image: []*image.Paletted{}, Delay: []int{}}

	/*
	Iterating over each image of the input file
	Use WaitGroups to ensure that all routines finish before proceeding
	https://gobyexample.com/waitgroups
	*/
	var wg sync.WaitGroup
	for i:=0; i<len(imgData.Image); i++ {
		wg.Add(1)
		go dither(i, imgData.Image[i], palette, paletteSize, imgArr, &wg)
	}
	wg.Wait()

	for i:=0; i<len(imgData.Image); i++ {
		/*
		Add dithered image to output struct
		Append 0 as frame delay
		For some reason, the output gif still plays *slightly* slower
		*/
		anim.Image = append(anim.Image, imgArr[i])
		anim.Delay = append(anim.Delay, 0)
	}

	/*
	Create output file
	Encode frames and write to output file
	*/
	gifOut, _ := os.Create(os.Args[2])
	defer gifOut.Close()
	gif.EncodeAll(gifOut, anim)
	fmt.Println("Written to", os.Args[2])

}
