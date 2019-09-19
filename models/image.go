package models

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/denisbakhtin/panteon/config"
	"github.com/nfnt/resize"
)

//Image type contains image info
type Image struct {
	Model

	URL        string  `form:"url"`
	PreviewURL string  `form:"-" binding:"-"`
	ProductID  uint64  `form:"product_id"`
	Hash       string  `gorm:"-" form:"-"`
	Product    Product `gorm:"save_associations:false" binding:"-" form:"-"`
}

//BeforeSave gorm hook
func (i *Image) BeforeSave() error {
	//if error happens, log it and return nil, don't prevent image saving on fail
	if err := i.ApplyWatermark(); err != nil {
		log.Printf("Error applying watermark to image, ID: %d, err: %s", i.ID, err)
	}
	if err := i.CreatePreview(); err != nil {
		log.Printf("Error creating image preview, ID: %d, err: %s", i.ID, err)
	}
	return nil
}

//CreatePreview creates image thumbnail with maxDim max dimensions, stored in jpg format
func (i *Image) CreatePreview() error {
	//max X and Y image dimension for previews
	const maxX = 175
	const maxY = 200

	prts := strings.SplitAfter(i.URL, "/uploads/")
	if len(prts) != 2 {
		return fmt.Errorf("Wrong number of URL splits")
	}
	parts := strings.Split(prts[1], "/")
	relPath := ""
	for i := 0; i < len(parts)-1; i++ {
		relPath = path.Join(relPath, parts[i])
	}
	//filename
	fname := parts[len(parts)-1]
	//original image folder
	srcPath := path.Join(config.UploadsPath(), relPath)
	//preview folder
	previewPath := path.Join(config.UploadsPath(), relPath, "previews")
	if err := os.MkdirAll(previewPath, 0755); err != nil {
		return err
	}
	file, err := os.Open(filepath.Join(srcPath, fname))
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := decodeImage(file)
	if err != nil {
		return err
	}
	img = resize.Thumbnail(maxX, maxY, img, resize.Lanczos3)
	bounds := img.Bounds()
	offset := image.Pt((maxX-bounds.Dx())/2, (maxY-bounds.Dy())/2)
	b := image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: maxX, Y: maxY}}
	m := image.NewRGBA(b)
	draw.Draw(m, b, image.NewUniform(color.RGBA{255, 255, 255, 255}), image.ZP, draw.Src)
	draw.Draw(m, bounds.Add(offset), img, image.ZP, draw.Over)

	dst, err := os.Create(path.Join(previewPath, fname))
	if err != nil {
		return err
	}
	defer dst.Close()
	if err := jpeg.Encode(dst, m, &jpeg.Options{Quality: 80}); err != nil {
		return err
	}

	relURL := strings.ReplaceAll(relPath, string(filepath.Separator), "/")
	if len(relURL) > 0 {
		relURL = fmt.Sprintf("%s/previews/%s", relURL, fname)
	} else {
		relURL = fmt.Sprintf("previews/%s", fname)
	}
	i.PreviewURL = fmt.Sprintf("/public/uploads/%s", relURL)
	return nil
}

//ApplyWatermark applies watermark to the image
func (i *Image) ApplyWatermark() error {
	watermarkPath := path.Join(config.PublicPath(), "watermark.png")
	prts := strings.SplitAfter(i.URL, "/uploads/")
	if len(prts) != 2 {
		return fmt.Errorf("Wrong number of URL splits")
	}
	parts := strings.Split(prts[1], "/")
	relPath := ""
	for i := 0; i < len(parts)-1; i++ {
		relPath = path.Join(relPath, parts[i])
	}
	//filename
	fname := parts[len(parts)-1]
	//original image folder
	srcPath := path.Join(config.UploadsPath(), relPath)
	file, err := os.Open(path.Join(srcPath, fname))
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := decodeImage(file)
	if err != nil {
		return err
	}
	b := img.Bounds()
	wmb, err := os.Open(watermarkPath)
	if err != nil {
		return err
	}
	defer wmb.Close()
	watermark, err := png.Decode(wmb)
	if err != nil {
		return err
	}
	watermark = resize.Thumbnail(uint(b.Dx()), uint(b.Dy()), watermark, resize.Lanczos3)
	wbounds := watermark.Bounds()
	offset := image.Pt((b.Dx()-wbounds.Dx())/2, (b.Dy()-wbounds.Dy())/2)

	//create resulting image
	m := image.NewRGBA(b)
	draw.Draw(m, b, image.NewUniform(color.RGBA{255, 255, 255, 255}), image.ZP, draw.Src)
	draw.Draw(m, b, img, image.ZP, draw.Over)
	draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)

	file.Close()
	//rewrite the file
	dst, err := os.Create(path.Join(srcPath, strings.ReplaceAll(fname, filepath.Ext(fname), ".jpg")))
	if err != nil {
		return err
	}
	defer dst.Close()
	if err := jpeg.Encode(dst, m, &jpeg.Options{Quality: 80}); err != nil {
		return err
	}

	//store new file name in the URL
	i.URL = strings.ReplaceAll(i.URL, filepath.Ext(fname), ".jpg")
	return nil
}

func decodeImage(file *os.File) (image.Image, error) {
	ext := strings.ToLower(filepath.Ext(file.Name()))
	var img image.Image
	var err error
	switch ext {
	case ".jpg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	default:
		return nil, fmt.Errorf("Unsupported image extension: %s, file: %s", ext, file.Name())
	}
	return img, err
}
