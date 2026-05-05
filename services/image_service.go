package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	imglib "mangav5/internal/image"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type ImageService struct{}

func NewImageService() *ImageService {
	return &ImageService{}
}

type WebPConvertOptions struct {
	Quality        int  `json:"quality"`
	Lossless       bool `json:"lossless"`
	Effort         int  `json:"effort"`
	SmartSubsample bool `json:"smartSubsample"`

	ResizeWidth  int  `json:"resizeWidth"`
	ResizeHeight int  `json:"resizeHeight"`
	AllowUpscale bool `json:"allowUpscale"`

	DeleteSource bool `json:"deleteSource"`

	Concurrency  int  `json:"concurrency"`
	Overwrite    bool `json:"overwrite"`
	SkipExisting bool `json:"skipExisting"`
	StopOnError  bool `json:"stopOnError"`
}

type WebPBatchResult struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	Bytes      int64  `json:"bytes"`
	Error      string `json:"error,omitempty"`
}

func (s *ImageService) ConvertSingleToWebP(inputPath string, outputPath string, options *WebPConvertOptions) (string, error) {
	inputPath = strings.TrimSpace(inputPath)
	outputPath = strings.TrimSpace(outputPath)
	if inputPath == "" {
		return "", errors.New("empty input path")
	}
	if outputPath == "" {
		ext := filepath.Ext(inputPath)
		outputPath = strings.TrimSuffix(inputPath, ext) + ".webp"
	}

	opts := toWebPOptions(options)
	c := imglib.FileToWebP(inputPath, opts)
	if options != nil {
		c.ResizeAspect(options.ResizeWidth, options.ResizeHeight, options.AllowUpscale)
	}
	if err := c.SaveWebP(outputPath).Err(); err != nil {
		return "", err
	}

	if options != nil && options.DeleteSource {
		inClean := filepath.Clean(inputPath)
		outClean := filepath.Clean(outputPath)
		if !strings.EqualFold(inClean, outClean) {
			if st, statErr := os.Stat(outputPath); statErr == nil && st.Size() > 0 {
				if rmErr := os.Remove(inputPath); rmErr != nil {
					return outputPath, rmErr
				}
			}
		}
	}

	return outputPath, nil
}

func (s *ImageService) ConvertBatchToWebP(inputPaths []string, outputDir string, options *WebPConvertOptions) ([]WebPBatchResult, error) {
	if len(inputPaths) == 0 {
		return []WebPBatchResult{}, nil
	}

	var ctx context.Context = context.Background()
	if app := application.Get(); app != nil {
		ctx = app.Context()
	}

	outputDir = strings.TrimSpace(outputDir)
	var items []imglib.WebPBatchItem
	if outputDir == "" {
		items = imglib.WebPBatchItemsFromPaths(inputPaths)
	} else {
		items = imglib.WebPBatchItemsFromPathsToDir(inputPaths, outputDir)
	}

	batch := imglib.WebPBatchOptions{}
	if options != nil {
		batch.Concurrency = options.Concurrency
		batch.Overwrite = options.Overwrite
		batch.SkipExisting = options.SkipExisting
		batch.StopOnError = options.StopOnError
		batch.DeleteSource = options.DeleteSource
		batch.ResizeWidth = options.ResizeWidth
		batch.ResizeHeight = options.ResizeHeight
		batch.AllowUpscale = options.AllowUpscale
	}

	results, _ := imglib.ConvertFilesToWebP(ctx, items, toWebPOptions(options), batch)
	out := make([]WebPBatchResult, 0, len(results))
	for _, r := range results {
		br := WebPBatchResult{
			InputPath:  r.Item.InputPath,
			OutputPath: r.Item.OutputPath,
			Bytes:      r.Bytes,
		}
		if r.Err != nil {
			br.Error = r.Err.Error()
		}
		out = append(out, br)
	}
	return out, nil
}

func toWebPOptions(options *WebPConvertOptions) imglib.WebPOptions {
	var out imglib.WebPOptions
	if options == nil {
		return out
	}
	out.Quality = options.Quality
	out.Lossless = options.Lossless
	out.Effort = options.Effort
	out.SmartSubsample = options.SmartSubsample
	return out
}

