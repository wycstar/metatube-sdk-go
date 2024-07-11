package engine

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/metatube-community/metatube-sdk-go/common/number"
	R "github.com/metatube-community/metatube-sdk-go/constant"
	"github.com/metatube-community/metatube-sdk-go/imageutil"
	"github.com/metatube-community/metatube-sdk-go/imageutil/pigo"
	"github.com/metatube-community/metatube-sdk-go/model"
	mt "github.com/metatube-community/metatube-sdk-go/provider"
)

// Default position constants for different kind of images.
const (
	defaultActorPrimaryImagePosition  = 0.5
	defaultMoviePrimaryImagePosition  = 1.0
	defaultMovieThumbImagePosition    = 0.5
	defaultMovieBackdropImagePosition = 0.0
)

func (e *Engine) GetActorPrimaryImage(name, id string) (image.Image, error) {
	info, err := e.GetActorInfoByProviderID(name, id, true)
	if err != nil {
		return nil, err
	}
	if len(info.Images) == 0 {
		return nil, mt.ErrImageNotFound
	}
	return e.GetImageByURL(e.MustGetActorProviderByName(name), info.Images[0], R.PrimaryImageRatio, defaultActorPrimaryImagePosition, false)
}

func (e *Engine) GetMoviePrimaryImage(name, id string, ratio, pos float64) (image.Image, error) {
	url, info, err := e.getPreferredMovieImageURLAndInfo(name, id, true)
	if err != nil {
		return nil, err
	}
	if ratio < 0 /* default primary ratio */ {
		ratio = R.PrimaryImageRatio
	}
	var auto bool
	if pos < 0 /* manual position disabled */ {
		pos = defaultMoviePrimaryImagePosition
		auto = number.RequireFaceDetection(info.Number)
	}
	return e.GetImageByURL(e.MustGetMovieProviderByName(name), url, ratio, pos, auto)
}

func (e *Engine) GetMovieThumbImage(name, id string) (image.Image, error) {
	url, _, err := e.getPreferredMovieImageURLAndInfo(name, id, false)
	if err != nil {
		return nil, err
	}
	return e.GetImageByURL(e.MustGetMovieProviderByName(name), url, R.ThumbImageRatio, defaultMovieThumbImagePosition, false)
}

func (e *Engine) GetMovieBackdropImage(name, id string) (image.Image, error) {
	url, _, err := e.getPreferredMovieImageURLAndInfo(name, id, false)
	if err != nil {
		return nil, err
	}
	return e.GetImageByURL(e.MustGetMovieProviderByName(name), url, R.BackdropImageRatio, defaultMovieBackdropImagePosition, false)
}

func (e *Engine) getImageIDByURL(url string) string {
	hash := md5.Sum([]byte(url))
	urlHash := hex.EncodeToString(hash[:])
	return urlHash
}

func (e *Engine) GetImageByURL(provider mt.Provider, url string, ratio, pos float64, auto bool) (img image.Image, err error) {
	if img, err = e.getImageByURL(provider, url); err != nil {
		return
	}
	if auto {
		start := time.Now()
		urlHash := e.getImageIDByURL(url)
		pos = pigo.CalculatePosition(img, ratio, pos, urlHash)
		duration := time.Since(start)
		fmt.Printf("%s %s %s\n", url, "CalculatePosition执行时间:", duration)
	}
	start := time.Now()
	img = imageutil.CropImagePosition(img, ratio, pos)
	duration := time.Since(start)
	fmt.Printf("%s %s %s\n", url, "CropImagePosition执行时间:", duration)
	return
}

func (e *Engine) saveImgToFile(img image.Image, filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("无法创建图像文件: %w", err)
	}
	defer file.Close()

	err = jpeg.Encode(file, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return fmt.Errorf("无法编码图像为 JPEG: %w", err)
	}

	return nil
}

func (e *Engine) getImageByURL(provider mt.Provider, url string) (img image.Image, err error) {
	currentDir := "d:/package/metatube/cache"
	urlHash := e.getImageIDByURL(url)
	imageFn := urlHash + ".jpg"
	imageCacheBasePath := filepath.Join(currentDir, "image_cache")
	_, err = os.Stat(imageCacheBasePath)
	if os.IsNotExist(err) {
		os.Mkdir(imageCacheBasePath, 0755)
	}
	imageCachePath := filepath.Join(imageCacheBasePath, imageFn)
	_, err = os.Stat(imageCachePath)
	if os.IsNotExist(err) {
		var resp *http.Response
		resp, err = e.Fetch(url, provider)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		img, _, err = image.Decode(resp.Body)
		e.saveImgToFile(img, imageCachePath)
		fmt.Printf("%s%s%s\n", url, "已保存到", imageCachePath)
		return
	} else {
		var file *os.File
		file, err = os.Open(imageCachePath)
		if err != nil {
			return
		}
		defer file.Close()
		img, _, err = image.Decode(file)
		fmt.Printf("%s%s%s\n", url, "缓存命中", imageCachePath)
		return
	}
}

func (e *Engine) getPreferredMovieImageURLAndInfo(name, id string, thumb bool) (url string, info *model.MovieInfo, err error) {
	info, err = e.GetMovieInfoByProviderID(name, id, true)
	if err != nil {
		return
	}
	url = info.CoverURL
	if thumb && info.BigThumbURL != "" /* big thumb > cover */ {
		url = info.BigThumbURL
	} else if !thumb && info.BigCoverURL != "" /* big cover > cover */ {
		url = info.BigCoverURL
	}
	return
}
