package cache

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/saaste/opengraph-image-creator/internal/pkg/config"
)

type CachedImage struct {
	Data    []byte
	ModTime time.Time
}

type ImageType int

const (
	ImageTypePng ImageType = iota
	ImageTypeJpeg
)

func TryGetImageFromCache(appConf *config.AppConfig, key string, imageType ImageType) (*CachedImage, error) {
	err := createCacheDirIfNotExist(appConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create the cache directory: %w", err)
	}

	var imageName string
	switch imageType {
	case ImageTypePng:
		imageName = fmt.Sprintf("%s.png", key)
	case ImageTypeJpeg:
		imageName = fmt.Sprintf("%s.jpg", key)
	default:
		return nil, fmt.Errorf("invalid image type")
	}

	filePath := fmt.Sprintf("%s/%s", appConf.CacheDir, imageName)

	stats, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Image %s not found in cache.\n", imageName)
			return nil, nil
		}
		return nil, fmt.Errorf("failed checking if image exists: %w", err)
	}

	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the image file: %w", err)
	}

	log.Printf("Image %s found in cache.\n", imageName)
	return &CachedImage{
		Data:    imageData,
		ModTime: stats.ModTime(),
	}, nil
}

func SaveImageToCache(appConf *config.AppConfig, key string, imageData []byte, imageType ImageType) (*CachedImage, error) {
	err := createCacheDirIfNotExist(appConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create the cache directory: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot as PNG image: %w", err)
	}

	var filePath string
	switch imageType {
	case ImageTypePng:
		filePath = fmt.Sprintf("%s/%s.png", appConf.CacheDir, key)
	case ImageTypeJpeg:
		filePath = fmt.Sprintf("%s/%s.jpg", appConf.CacheDir, key)
	default:
		return nil, fmt.Errorf("invalid image type")
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create the output file: %w", err)
	}
	defer outFile.Close()

	switch imageType {
	case ImageTypePng:
		encoder := png.Encoder{
			CompressionLevel: png.BestCompression,
		}
		if err = encoder.Encode(outFile, img); err != nil {
			return nil, fmt.Errorf("failed to encode PNG image: %w", err)
		}
	case ImageTypeJpeg:
		options := jpeg.Options{
			Quality: appConf.JpegCompression,
		}
		if err = jpeg.Encode(outFile, img, &options); err != nil {
			return nil, fmt.Errorf("failed to encode JPEG image: %w", err)
		}

	default:
		return nil, fmt.Errorf("invalid image type")
	}

	return TryGetImageFromCache(appConf, key, imageType)
}

func createCacheDirIfNotExist(appConf *config.AppConfig) error {
	if _, err := os.Stat(appConf.CacheDir); os.IsNotExist(err) {
		log.Println("Cache dir does not exist. Creating...")
		if err = os.MkdirAll(appConf.CacheDir, os.ModePerm); err != nil {
			return err
		}
		log.Println("Cache dir created.")
	}
	return nil
}
