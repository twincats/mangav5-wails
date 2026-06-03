package image

import (
	"context"
	"errors"
	"image"
	"image/draw"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/cshum/vipsgen/vips"
)

var startupOnce sync.Once
var shutdownOnce sync.Once

// Startup libvips sekali saja
func Startup() {
	startupOnce.Do(func() {
		vips.Startup(nil)
	})
}

// Shutdown libvips, panggil saat app selesai kalau perlu.
// Aman dipanggil berkali-kali.
func Shutdown() {
	shutdownOnce.Do(func() {
		vips.Shutdown()
	})
}

// WebPOptions adalah opsi yang dipakai untuk export WebP.
type WebPOptions struct {
	Quality        int // 1-100
	Lossless       bool
	Effort         int // 0-6
	SmartSubsample bool
}

// normalize memberikan default value kalau user tidak isi.
func (o WebPOptions) normalize() WebPOptions {
	if o.Quality <= 0 {
		o.Quality = 80
	}
	if o.Quality > 100 {
		o.Quality = 100
	}
	if o.Effort < 0 {
		o.Effort = 0
	}
	if o.Effort > 6 {
		o.Effort = 6
	}
	return o
}

// WebPChain menyimpan state untuk method chaining.
type WebPChain struct {
	img    *vips.Image
	source *vips.Source
	err    error
	opts   WebPOptions
	closed bool
}

// NewWebPChain membuat chain baru.
func NewWebPChain(opts WebPOptions) *WebPChain {
	Startup()
	return &WebPChain{
		opts: opts.normalize(),
	}
}

// --------------------------------------------------
// Convenience constructor
// --------------------------------------------------

func FileToWebP(path string, opts WebPOptions) *WebPChain {
	return NewWebPChain(opts).FromFile(path)
}

func BufferToWebP(buf []byte, opts WebPOptions) *WebPChain {
	return NewWebPChain(opts).FromBuffer(buf)
}

func ReaderToWebP(r io.Reader, opts WebPOptions) *WebPChain {
	return NewWebPChain(opts).FromReader(r)
}

func GoImageToWebP(img image.Image, opts WebPOptions) *WebPChain {
	return NewWebPChain(opts).FromImage(img)
}

// --------------------------------------------------
// Input loaders
// --------------------------------------------------

// FromFile load image dari file path.
func (c *WebPChain) FromFile(path string) *WebPChain {
	if c.err != nil {
		return c
	}
	if path == "" {
		c.err = errors.New("empty file path")
		return c
	}

	img, err := vips.NewImageFromFile(path, nil)
	if err != nil {
		c.err = err
		return c
	}

	c.img = img
	return c
}

// FromBuffer load image dari []byte.
func (c *WebPChain) FromBuffer(buf []byte) *WebPChain {
	if c.err != nil {
		return c
	}
	if len(buf) == 0 {
		c.err = errors.New("empty buffer")
		return c
	}

	img, err := vips.NewImageFromBuffer(buf, nil)
	if err != nil {
		c.err = err
		return c
	}

	c.img = img
	return c
}

// FromReader load image dari io.Reader.
// Karena vips.NewSource butuh io.ReadCloser, kita bungkus jika perlu.
// Source harus tetap hidup selama image masih dipakai.
func (c *WebPChain) FromReader(r io.Reader) *WebPChain {
	if c.err != nil {
		return c
	}
	if r == nil {
		c.err = errors.New("nil reader")
		return c
	}

	var rc io.ReadCloser
	if r2, ok := r.(io.ReadCloser); ok {
		rc = r2
	} else {
		rc = io.NopCloser(r)
	}

	src := vips.NewSource(rc)
	img, err := vips.NewImageFromSource(src, nil)
	if err != nil {
		src.Close()
		c.err = err
		return c
	}

	c.source = src
	c.img = img
	return c
}

// FromImage load dari Go image.Image.
// Kita ubah dulu ke NRGBA lalu kirim raw pixels ke NewImageFromMemory.
func (c *WebPChain) FromImage(src image.Image) *WebPChain {
	if c.err != nil {
		return c
	}
	if src == nil {
		c.err = errors.New("nil image")
		return c
	}

	nrgba, raw := toContiguousNRGBA(src)
	w := nrgba.Bounds().Dx()
	h := nrgba.Bounds().Dy()

	img, err := vips.NewImageFromMemory(raw, w, h, 4)
	if err != nil {
		c.err = err
		return c
	}

	c.img = img
	return c
}

func (c *WebPChain) ResizeAspect(targetWidth int, targetHeight int, allowUpscale bool) *WebPChain {
	if c.err != nil {
		return c
	}
	if c.img == nil {
		c.err = errors.New("no image loaded")
		return c
	}
	if targetWidth <= 0 && targetHeight <= 0 {
		return c
	}
	if targetWidth > 0 && targetHeight > 0 {
		c.err = errors.New("set either targetWidth or targetHeight")
		return c
	}

	w := c.img.Width()
	h := c.img.Height()
	if w <= 0 || h <= 0 {
		c.err = errors.New("invalid source image size")
		return c
	}

	var scale float64
	if targetWidth > 0 {
		scale = float64(targetWidth) / float64(w)
	} else {
		scale = float64(targetHeight) / float64(h)
	}
	if scale <= 0 {
		c.err = errors.New("invalid scale ratio")
		return c
	}
	if !allowUpscale && scale > 1 {
		return c
	}
	if scale == 1 {
		return c
	}

	c.err = c.img.Resize(scale, nil)
	return c
}

// --------------------------------------------------
// Output methods (terminal)
// --------------------------------------------------

// SaveWebP save langsung ke file dan tetap return chain supaya bisa .Err()
func (c *WebPChain) SaveWebP(path string) *WebPChain {
	defer c.release()

	if c.err != nil {
		return c
	}
	if c.img == nil {
		c.err = errors.New("no image loaded")
		return c
	}
	if path == "" {
		c.err = errors.New("empty output path")
		return c
	}

	c.err = c.img.Webpsave(path, c.toWebpsaveOptions())
	return c
}

func (c *WebPChain) WriteWebPTo(w io.Writer) *WebPChain {
	defer c.release()

	if c.err != nil {
		return c
	}
	if c.img == nil {
		c.err = errors.New("no image loaded")
		return c
	}
	if w == nil {
		c.err = errors.New("nil writer")
		return c
	}

	target := vips.NewTarget(nopWriteCloser{Writer: w})
	defer target.Close()

	c.err = c.img.WebpsaveTarget(target, c.toWebpsaveTargetOptions())
	return c
}

func (c *WebPChain) WriteTo(w io.Writer) (int64, error) {
	defer c.release()

	if c.err != nil {
		return 0, c.err
	}
	if c.img == nil {
		return 0, errors.New("no image loaded")
	}
	if w == nil {
		return 0, errors.New("nil writer")
	}

	cw := &countingWriter{Writer: w}
	target := vips.NewTarget(nopWriteCloser{Writer: cw})
	defer target.Close()

	if err := c.img.WebpsaveTarget(target, c.toWebpsaveTargetOptions()); err != nil {
		return cw.n, err
	}
	return cw.n, nil
}

// Bytes export hasil WebP menjadi []byte.
// Ini terminal method.
func (c *WebPChain) Bytes() ([]byte, error) {
	defer c.release()

	if c.err != nil {
		return nil, c.err
	}
	if c.img == nil {
		return nil, errors.New("no image loaded")
	}

	buf, err := c.img.WebpsaveBuffer(c.toWebpsaveBufferOptions())
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Err mengambil error terakhir dari chain.
func (c *WebPChain) Err() error {
	return c.err
}

// Close kalau user mau stop sebelum terminal method dipanggil.
func (c *WebPChain) Close() {
	c.release()
}

// --------------------------------------------------
// Helpers
// --------------------------------------------------

func (c *WebPChain) release() {
	if c.closed {
		return
	}
	c.closed = true

	if c.img != nil {
		c.img.Close()
		c.img = nil
	}
	if c.source != nil {
		c.source.Close()
		c.source = nil
	}
}

func (c *WebPChain) toWebpsaveOptions() *vips.WebpsaveOptions {
	return &vips.WebpsaveOptions{
		Q:              c.opts.Quality,
		Lossless:       c.opts.Lossless,
		Effort:         c.opts.Effort,
		SmartSubsample: c.opts.SmartSubsample,
	}
}

func (c *WebPChain) toWebpsaveBufferOptions() *vips.WebpsaveBufferOptions {
	return &vips.WebpsaveBufferOptions{
		Q:              c.opts.Quality,
		Lossless:       c.opts.Lossless,
		Effort:         c.opts.Effort,
		SmartSubsample: c.opts.SmartSubsample,
	}
}

func (c *WebPChain) toWebpsaveTargetOptions() *vips.WebpsaveTargetOptions {
	return &vips.WebpsaveTargetOptions{
		Q:              c.opts.Quality,
		Lossless:       c.opts.Lossless,
		Effort:         c.opts.Effort,
		SmartSubsample: c.opts.SmartSubsample,
	}
}

// toContiguousNRGBA mengubah image.Image apapun ke *image.NRGBA
// lalu memastikan raw bytes-nya kontigu (tanpa stride ekstra).
func toContiguousNRGBA(src image.Image) (*image.NRGBA, []byte) {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)

	rowBytes := b.Dx() * 4
	if dst.Stride == rowBytes {
		return dst, dst.Pix
	}

	raw := make([]byte, rowBytes*b.Dy())
	for y := 0; y < b.Dy(); y++ {
		srcStart := y * dst.Stride
		srcEnd := srcStart + rowBytes
		dstStart := y * rowBytes
		copy(raw[dstStart:dstStart+rowBytes], dst.Pix[srcStart:srcEnd])
	}
	return dst, raw
}

type nopWriteCloser struct {
	io.Writer
}

func (n nopWriteCloser) Close() error {
	return nil
}

type countingWriter struct {
	io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.Writer.Write(p)
	c.n += int64(n)
	return n, err
}

type WebPBatchItem struct {
	InputPath  string
	OutputPath string
}

type WebPBatchOptions struct {
	Concurrency  int
	Overwrite    bool
	SkipExisting bool
	StopOnError  bool
	DeleteSource bool
	ResizeWidth  int
	ResizeHeight int
	AllowUpscale bool
}

type WebPBatchResult struct {
	Item  WebPBatchItem
	Bytes int64
	Err   error
}

func WebPBatchItemsFromPaths(inputPaths []string) []WebPBatchItem {
	items := make([]WebPBatchItem, 0, len(inputPaths))
	for _, in := range inputPaths {
		in = strings.TrimSpace(in)
		if in == "" {
			continue
		}
		ext := filepath.Ext(in)
		out := strings.TrimSuffix(in, ext) + ".webp"
		items = append(items, WebPBatchItem{
			InputPath:  in,
			OutputPath: out,
		})
	}
	return items
}

func WebPBatchItemsFromPathsToDir(inputPaths []string, outputDir string) []WebPBatchItem {
	outputDir = strings.TrimSpace(outputDir)
	items := make([]WebPBatchItem, 0, len(inputPaths))
	for _, in := range inputPaths {
		in = strings.TrimSpace(in)
		if in == "" {
			continue
		}
		base := filepath.Base(in)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext) + ".webp"
		out := filepath.Join(outputDir, name)
		items = append(items, WebPBatchItem{
			InputPath:  in,
			OutputPath: out,
		})
	}
	return items
}

func ConvertFilesToWebP(ctx context.Context, items []WebPBatchItem, opts WebPOptions, batch WebPBatchOptions) ([]WebPBatchResult, error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	if len(items) == 0 {
		return []WebPBatchResult{}, nil
	}

	workers := batch.Concurrency
	if workers <= 0 {
		workers = defaultBatchConcurrency()
	}
	if workers > len(items) {
		workers = len(items)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type job struct {
		idx  int
		item WebPBatchItem
	}
	type res struct {
		idx int
		r   WebPBatchResult
	}

	jobs := make(chan job)
	results := make(chan res)

	var wg sync.WaitGroup
	workerFn := func() {
		defer wg.Done()
		for j := range jobs {
			if err := ctx.Err(); err != nil {
				return
			}

			r := WebPBatchResult{Item: j.item}
			if j.item.InputPath == "" || j.item.OutputPath == "" {
				r.Err = errors.New("empty input/output path")
				results <- res{idx: j.idx, r: r}
				if batch.StopOnError {
					cancel()
				}
				continue
			}

			if !batch.Overwrite || batch.SkipExisting {
				if _, err := os.Stat(j.item.OutputPath); err == nil {
					if batch.SkipExisting {
						results <- res{idx: j.idx, r: r}
						continue
					}
					r.Err = errors.New("output already exists")
					results <- res{idx: j.idx, r: r}
					if batch.StopOnError {
						cancel()
					}
					continue
				}
			}

			if dir := filepath.Dir(j.item.OutputPath); dir != "" && dir != "." {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					r.Err = err
					results <- res{idx: j.idx, r: r}
					if batch.StopOnError {
						cancel()
					}
					continue
				}
			}

			c := FileToWebP(j.item.InputPath, opts)
			targetW := batch.ResizeWidth
			targetH := batch.ResizeHeight
			if targetW > 0 && targetH == 0 && c.img != nil {
				w := c.img.Width()
				h := c.img.Height()
				if w > 0 && h > 0 && w > h {
					const landscapeMinWidth = 1980
					if w <= landscapeMinWidth {
						targetW = 0
						targetH = 0
					} else {
						targetW = landscapeMinWidth
					}
				}
			}
			c.ResizeAspect(targetW, targetH, batch.AllowUpscale)
			n, err := func() (int64, error) {
				f, err := os.Create(j.item.OutputPath)
				if err != nil {
					return 0, err
				}
				defer f.Close()
				return c.WriteTo(f)
			}()
			r.Bytes = n
			if err != nil {
				r.Err = err
			} else if batch.DeleteSource {
				inClean := filepath.Clean(j.item.InputPath)
				outClean := filepath.Clean(j.item.OutputPath)
				if !strings.EqualFold(inClean, outClean) {
					if st, statErr := os.Stat(j.item.OutputPath); statErr == nil && st.Size() > 0 {
						if rmErr := os.Remove(j.item.InputPath); rmErr != nil {
							r.Err = rmErr
						}
					}
				}
			}
			results <- res{idx: j.idx, r: r}
			if batch.StopOnError && r.Err != nil {
				cancel()
				return
			}
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go workerFn()
	}

	go func() {
		defer close(results)
		wg.Wait()
	}()

	go func() {
		defer close(jobs)
		for i, it := range items {
			if err := ctx.Err(); err != nil {
				return
			}
			jobs <- job{idx: i, item: it}
		}
	}()

	out := make([]WebPBatchResult, len(items))
	var joinedErr error
	for i := 0; i < len(items); i++ {
		select {
		case <-ctx.Done():
			for j := i; j < len(items); j++ {
				out[j] = WebPBatchResult{Item: items[j], Err: ctx.Err()}
			}
			if joinedErr == nil {
				joinedErr = ctx.Err()
			} else {
				joinedErr = errors.Join(joinedErr, ctx.Err())
			}
			return out, joinedErr
		case rr, ok := <-results:
			if !ok {
				return out, joinedErr
			}
			out[rr.idx] = rr.r
			if rr.r.Err != nil {
				if joinedErr == nil {
					joinedErr = rr.r.Err
				} else {
					joinedErr = errors.Join(joinedErr, rr.r.Err)
				}
			}
		}
	}

	return out, joinedErr
}

func defaultBatchConcurrency() int {
	n := runtime.GOMAXPROCS(0)
	if n < 1 {
		n = 1
	}
	if n > 4 {
		n = 4
	}
	return n
}

// --------------------------------------------------
// Optional utility shortcut
// --------------------------------------------------

// SaveBytesToFile utility helper kalau kamu pakai .Bytes()
func SaveBytesToFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
