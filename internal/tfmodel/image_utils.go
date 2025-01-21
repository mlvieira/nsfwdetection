package tfmodel

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/wamuir/graft/tensorflow"
)

// openImage Opens and decode an image
func openImage(filePath string) (image.Image, string, error) {
	ext := filepath.Ext(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	if ext == ".webp" {
		img, err := handleWebPFrame(filePath, file)
		if err != nil {
			return nil, "webp", fmt.Errorf("error processing WebP file: %w", err)
		}
		return img, "webp", nil
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("error decoding file: %w", err)
	}

	return img, format, nil
}

// handleAnimatedFrames Handles animated images (GIF, WebP)
func handleAnimatedFrames(filePath string, img image.Image, format string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reopening file: %w", err)
	}
	defer file.Close()

	var finalFrame image.Image = img

	if format == "gif" {
		finalFrame, err = processGIF(file)
		if err != nil {
			return nil, err
		}
	}

	return finalFrame, nil
}

// processGIF decodes a multi-frame GIF and returns a single composited image.
func processGIF(file *os.File) (image.Image, error) {
	gifData, err := gif.DecodeAll(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding GIF frames: %w", err)
	}

	if len(gifData.Image) <= 1 {
		return gifData.Image[0], nil
	}

	midFrame := len(gifData.Image) / 2

	finalFrame := composeGIFFrame(gifData, midFrame)
	return finalFrame, nil
}

// composeGIFFrame combines multiple GIF frames into a single composited image.
func composeGIFFrame(gifData *gif.GIF, frameIndex int) image.Image {
	bounds := gifData.Image[0].Bounds()
	composed := image.NewRGBA(bounds)

	if gifData.BackgroundIndex < uint8(len(gifData.Image[0].Palette)) {
		bgColor := gifData.Image[0].Palette[gifData.BackgroundIndex]
		draw.Draw(composed, bounds, &image.Uniform{C: bgColor}, image.Point{}, draw.Src)
	}

	for i := 0; i <= frameIndex; i++ {
		frame := gifData.Image[i]

		switch gifData.Disposal[i] {
		case gif.DisposalBackground:
			draw.Draw(composed, frame.Bounds(), &image.Uniform{C: gifData.Image[0].Palette[gifData.BackgroundIndex]}, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
		default:
			draw.Draw(composed, frame.Bounds(), frame, image.Point{}, draw.Over)
		}
	}

	return composed
}

// resizeAndNormalize Resizes and normalize image
// this is mostly from the readme
// https://github.com/bhky/opennsfw2/tree/main?tab=readme-ov-file#preprocessing-details
func resizeAndNormalize(img image.Image) (*tensorflow.Tensor, error) {
	const targetSize = 224

	// resize image to 256 x 256
	resized := imaging.Resize(img, 256, 256, imaging.Lanczos)

	// store the image as jpeg in memory and reload it again
	img, err := reloadImage(resized)
	if err != nil {
		return nil, fmt.Errorf("error reloading image: %w", err)
	}

	// crop the centre part with size 224 x 224
	cropped := imaging.CropCenter(img, targetSize, targetSize)

	var tensorData [1][224][224][3]float32

	if cropped.Bounds().Dx() != 224 || cropped.Bounds().Dy() != 224 {
		return nil, fmt.Errorf("resized image has invalid dimensions: %dx%d", cropped.Bounds().Dx(), cropped.Bounds().Dy())
	}

	for y := 0; y < targetSize; y++ {
		for x := 0; x < targetSize; x++ {
			r, g, b := extractRGB(cropped.At(x, y))

			// swap the color channels to bgr
			// we need to convert 16 bit to 8 bit  due to python funkyness
			// then substract mean value of each channel
			tensorData[0][y][x][0] = float32(b) - 104.0
			tensorData[0][y][x][1] = float32(g) - 117.0
			tensorData[0][y][x][2] = float32(r) - 123.0
		}
	}

	tensor, err := tensorflow.NewTensor(tensorData)
	if err != nil {
		return nil, fmt.Errorf("failed to create tensor: %w", err)
	}
	return tensor, nil
}

// reloadImage reloads the image as jpeg (some yahoo shit)
func reloadImage(img image.Image) (image.Image, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return nil, err
	}

	reloadedImg, _, err := image.Decode(&buf)
	if err != nil {
		return nil, err
	}
	return reloadedImg, nil
}

// handleWebPFrame processes WebP frames
func handleWebPFrame(filePath string, file *os.File) (image.Image, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading WebP data: %w", err)
	}

	// animated
	if len(data) > 20 && (data[20]&0x02) != 0 {
		return extractFrameImageMagick(filePath, -1)
	}

	decodedImage, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("error decoding WebP: %w", err)
	}

	return decodedImage, nil
}

// extractRGB extracts 8-bit RGB values from a color.Color.
func extractRGB(c color.Color) (uint8, uint8, uint8) {
	r, g, b, _ := c.RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)
}

// extractFrameImageMagick - Extracts the middle frame using ImageMagick
func extractFrameImageMagick(filePath string, frameIndex int) (image.Image, error) {
	frameArg := fmt.Sprintf("%s[%d]", filePath, frameIndex)

	cmd := exec.Command("convert", frameArg, "png:-")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error extracting frame with ImageMagick: %w", err)
	}

	img, _, err := image.Decode(&out)
	if err != nil {
		return nil, fmt.Errorf("error decoding extracted frame: %w", err)
	}
	return img, nil
}
