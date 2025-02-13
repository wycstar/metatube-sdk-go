package pigo

import (
	"encoding/gob"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"time"

	pigo "github.com/esimov/pigo/core"
)

var classifier *pigo.Pigo

func init() {
	classifier, _ = pigo.NewPigo().Unpack(cascade)
}

func loadCascadeResult(savePath string, dets *[]pigo.Detection) error {
	file, err := os.Open(savePath)
	if err != nil {
		return fmt.Errorf("无法打开缓存文件: %w", err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(dets)
	if err != nil {
		return fmt.Errorf("无法解码结果: %w", err)
	}

	return nil
}

func saveCascadeResult(savePath string, dets []pigo.Detection) error {
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("无法创建缓存文件: %w", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(dets)
	if err != nil {
		return fmt.Errorf("无法编码结果: %w", err)
	}

	return nil
}

func DetectFaces(img image.Image, imageID string) (dets []pigo.Detection) {
	cParams := pigo.CascadeParams{
		MinSize:     20,
		MaxSize:     2000,
		ShiftFactor: 0.1,
		ScaleFactor: 1.1,
		ImageParams: pigo.ImageParams{
			Pixels: pigo.RgbToGrayscale(img),
			Rows:   img.Bounds().Dy(),
			Cols:   img.Bounds().Dx(),
			Dim:    img.Bounds().Dx(),
		},
	}
	currentDir := "/home/wyc/package/metatube/cache"
	casCacheBasePath := filepath.Join(currentDir, "face_cache")
	_, err := os.Stat(casCacheBasePath)
	if os.IsNotExist(err) {
		os.Mkdir(casCacheBasePath, 0755)
	}
	casCachePath := filepath.Join(casCacheBasePath, imageID+".gob")
	if _, err := os.Stat(casCachePath); err == nil {
		// 加载缓存结果
		if err = loadCascadeResult(casCachePath, &dets); err != nil {
			return
		}
		fmt.Printf("%s %s\n", casCachePath, "命中dets缓存")
	} else {
		start := time.Now()
		// Run the classifier over the obtained leaf nodes and return the detection results.
		// The result contains quadruplets representing the row, column, scale and detection score.
		dets = classifier.RunCascade(cParams, 0.0)
		duration := time.Since(start)
		fmt.Printf("%s %s\n", "RunCascade执行时间:", duration)
		saveCascadeResult(casCachePath, dets)
	}
	// Calculate the intersection over union (IoU) of two clusters.
	dets = classifier.ClusterDetections(dets, 0.2)
	return
}

func CalculatePosition(img image.Image, ratio float64, pos float64, imageID string) float64 {
	if dets := DetectFaces(img, imageID); len(dets) > 0 {
		sort.SliceStable(dets, func(i, j int) bool {
			return float32(dets[i].Scale)*dets[i].Q > float32(dets[j].Scale)*dets[j].Q
		})
		var (
			width  = img.Bounds().Dx()
			height = img.Bounds().Dy()
		)
		if int(float64(height)*ratio) < width {
			pos = float64(dets[0].Col) / float64(width)
		} else {
			pos = float64(dets[0].Row) / float64(height)
		}
	}
	return pos
}
