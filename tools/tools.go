package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/denisbakhtin/panteon/config"
	"github.com/denisbakhtin/panteon/models"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Task not specified")
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	gin.SetMode("debug")
	config.LoadConfig()
	models.SetDB(config.GetConnectionString())

	task := os.Args[1]
	switch task {
	case "importgrav":
		if len(os.Args) < 4 {
			log.Fatal("Usage: go run tools/tools.go importgrav inputPath watermarkPath")
		}
		//inputPath should be a directory with subdirectories, each containing images and texts.txt with product names (1 each line)
		//the subdirectory name goes to the category name being created
		importProducts(os.Args[2], os.Args[3])
	case "cureproducts":
		cureProducts()
	case "cureimages":
		cureImages()
	case "importcat":
		if len(os.Args) < 7 {
			log.Fatal("Usage: go run tools/tools.go importcat inputPath watermarkPath categoryID productName nameToCode{0, 1}")
		}
		importCategory(os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6])
	case "makepreviews":
		makeImagePreviews(len(os.Args) == 3 && os.Args[2] == "force")
	default:
		log.Fatal("Unknown task")
	}
}

func importProducts(inputPath, watermarkPath string) {
	log.Println("Starting importProducts")
	subdirs, err := ioutil.ReadDir(inputPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range subdirs {
		if !f.IsDir() {
			continue
		}
		category := models.Category{Title: title(f.Name()), Published: true, Products: make([]models.Product, 0, 200)}
		files, err := ioutil.ReadDir(filepath.Join(inputPath, f.Name()))
		if err != nil {
			log.Fatal(err)
		}

		pfile, err := os.Open(filepath.Join(inputPath, f.Name(), "texts.txt"))
		if err != nil {
			log.Fatal(err)
		}
		defer pfile.Close()

		// Start reading from the file with a reader.
		preader := bufio.NewReader(pfile)

		for _, img := range files {
			//move to closure, use defer file close, check invalid file descriptor on line 86 error
			if !strings.HasSuffix(strings.ToLower(img.Name()), ".jpg") {
				continue
			}
			dstPath := filepath.Join(config.UploadsPath(), f.Name())
			os.MkdirAll(dstPath, 0755)
			func(preader *bufio.Reader) {
				productName, err := preader.ReadString('\n')
				if err == io.EOF {
					return
				}
				if err != nil {
					log.Fatalf("Error reading texts.txt: %s\n", err)
				}
				product := models.Product{Title: title(productName), Published: true}
				fullName := filepath.Join(dstPath, img.Name())

				src, err := os.Open(filepath.Join(inputPath, f.Name(), img.Name()))
				if err != nil {
					log.Fatal(err)
				}
				jsrc, _ := jpeg.Decode(src)
				defer src.Close()

				m := applyWatermark(jsrc, watermarkPath)
				dst, err := os.Create(fullName)
				if err != nil {
					log.Fatal(err)
				}
				defer dst.Close()
				if err := jpeg.Encode(dst, m, &jpeg.Options{Quality: 80}); err != nil {
					log.Fatal(err)
				}

				uri := fmt.Sprintf("/public/uploads/%s/%s", f.Name(), img.Name())
				image := models.Image{URL: uri}
				product.Images = []models.Image{image}
				category.Products = append(category.Products, product)
			}(preader)
		}
		log.Println("Saving category in db: ", category.Title, len(category.Products))
		if err := models.GetDB().Create(&category).Error; err != nil {
			log.Fatal(err)
		}
		log.Println("Category saved: ", category.Title, len(category.Products))
	}
}

//cureProducts applies proper case to products title, assigns default image id and may be more...
func cureProducts() {
	var products []models.Product
	models.GetDB().Preload("Images").Find(&products)
	for _, p := range products {
		oTitle := p.Title
		oDefaultImageID := p.DefaultImageID
		oSlug := p.Slug

		p.Slug = strings.ReplaceAll(p.Slug, "--", "-")
		p.Slug = strings.Trim(p.Slug, "-")
		p.Title = title(p.Title)
		if p.DefaultImageID == 0 && len(p.Images) > 0 {
			p.DefaultImageID = p.Images[0].ID
		}
		if oTitle != p.Title || oDefaultImageID != p.DefaultImageID || oSlug != p.Slug {
			models.GetDB().Model(&p).Updates(models.Product{Title: p.Title, DefaultImageID: p.DefaultImageID, Slug: p.Slug})
		}
	}
}

//cureImages fixes some issues with images after import
func cureImages() {
	var images []models.Image
	models.GetDB().Where("url like ?", "/home/%").Find(&images)
	for _, i := range images {
		parts := strings.SplitAfter(i.URL, "/home/tabula/denisbakhtin/panteon/")
		if len(parts) != 2 {
			log.Fatal("Wrong number of parts in image url, ID: ", i.ID)
		}
		i.URL = "/" + parts[1]
		models.GetDB().Model(&i).Updates(models.Image{URL: i.URL})
	}
	models.GetDB().Where("url like ?", "%//%").Find(&images)
	for _, i := range images {
		parts := strings.SplitAfter(i.URL, "//")
		if len(parts) != 2 {
			log.Fatal("Wrong number of parts in image url, ID: ", i.ID)
		}
		i.URL = "/" + parts[1]
		models.GetDB().Model(&i).Updates(models.Image{URL: i.URL})
	}
}

func title(s string) string {
	s = strings.ReplaceAll(s, "  ", " ")
	words := strings.Split(s, " ")
	if len(words) > 0 {
		words[0] = strings.Title(words[0])
		s = strings.Join(words, " ")
	}
	return strings.TrimSpace(s)
}

func importCategory(inputPath, watermarkPath, catID, productName, nameToCode string) {
	files, err := ioutil.ReadDir(inputPath)
	if err != nil {
		log.Fatal(err)
	}
	dstPath := filepath.Join(config.UploadsPath(), path.Base(inputPath))
	os.MkdirAll(dstPath, 0755)
	category := models.Category{}
	models.GetDB().First(&category, catID)
	runes := []rune(category.Title)
	codePref := strings.ToUpper(string(runes[0:2]))

	for _, f := range files {
		func() {
			file, err := os.Open(filepath.Join(inputPath, f.Name()))
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			img, err := decodeImage(file)
			if err != nil {
				log.Fatal(err)
			}
			m := applyWatermark(img, watermarkPath)
			fullName := filepath.Join(dstPath, strings.Replace(path.Base(file.Name()), filepath.Ext(file.Name()), ".jpg", -1))
			dst, err := os.Create(fullName)
			if err != nil {
				log.Fatal(err)
			}
			defer dst.Close()
			if err := jpeg.Encode(dst, m, &jpeg.Options{Quality: 80}); err != nil {
				log.Fatal(err)
			}

			product := models.Product{Title: productName, Published: true, CategoryID: atouint64(catID)}
			if nameToCode == "1" {
				product.Code = fmt.Sprintf("%s-%s", codePref, strings.Split(path.Base(dst.Name()), ".")[0])
			}
			uri := fmt.Sprintf("/public/uploads/%s/%s", path.Base(inputPath), path.Base(dst.Name()))
			image := models.Image{URL: uri}
			product.Images = []models.Image{image}
			if err := models.GetDB().Create(&product).Error; err != nil {
				log.Fatal(err)
			}
			models.GetDB().Model(&product).Update("default_image_id", product.Images[0].ID)
		}()
	}
}

func applyWatermark(img image.Image, watermarkPath string) image.Image {
	b := img.Bounds()
	wmb, err := os.Open(watermarkPath)
	if err != nil {
		log.Fatal(err)
	}
	defer wmb.Close()
	watermark, err := png.Decode(wmb)
	if err != nil {
		log.Fatal(err)
	}

	watermark = resize.Thumbnail(uint(b.Dx()), uint(b.Dy()), watermark, resize.Lanczos3)
	wbounds := watermark.Bounds()
	offset := image.Pt((b.Dx()-wbounds.Dx())/2, (b.Dy()-wbounds.Dy())/2)

	//create resulting image
	m := image.NewRGBA(b)
	draw.Draw(m, b, image.NewUniform(color.RGBA{255, 255, 255, 255}), image.ZP, draw.Src)
	draw.Draw(m, b, img, image.ZP, draw.Over)
	draw.Draw(m, watermark.Bounds().Add(offset), watermark, image.ZP, draw.Over)
	return m
}

//atouint64 converts string to uint64, returns 0 if error
func atouint64(s string) uint64 {
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

//makeImagePreviews walks all Images in db and creates image previews via Image.BeforeSave hook
func makeImagePreviews(force bool) {
	log.Println("This task will create previews for all images without preview :). To force recreation call Image.CreatePreview() for each.")
	var images []models.Image
	db := models.GetDB()
	dbq := db
	if !force {
		dbq = dbq.Where("preview_url is null or preview_url = ?", "")
	}
	dbq.Find(&images)
	for _, img := range images {
		if err := db.Save(&img).Error; err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Processed %d images. Finished.\n", len(images))
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
