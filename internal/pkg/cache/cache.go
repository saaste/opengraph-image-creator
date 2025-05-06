package cache

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/saaste/opengraph-image-creator/internal/pkg/config"
)

type CachedImage struct {
	Data    []byte
	ModTime time.Time
}

func TryGetImageFromCache(appConf *config.AppConfig, key string) (*CachedImage, error) {
	err := createCacheDirIfNotExist(appConf)
	if err != nil {
		return nil, fmt.Errorf("failed to create the cache directory: %w", err)
	}

	imageName := fmt.Sprintf("%s.png", key)
	filePath := fmt.Sprintf("%s/%s", appConf.CacheDir, imageName)

	stats, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Image %s not found in cache\n", imageName)
			return nil, nil
		}
		return nil, fmt.Errorf("failed checking if image exists: %w", err)
	}

	imageData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the image file: %w", err)
	}

	log.Printf("Image %s found in cache\n", imageName)
	return &CachedImage{
		Data:    imageData,
		ModTime: stats.ModTime(),
	}, nil
}

func SaveImageToCache(appConf *config.AppConfig, key string, imageData []byte) error {
	err := createCacheDirIfNotExist(appConf)
	if err != nil {
		return fmt.Errorf("failed to create the cache directory: %w", err)
	}

	filePath := fmt.Sprintf("%s/%s.png", appConf.CacheDir, key)
	return os.WriteFile(filePath, imageData, 0644)
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
