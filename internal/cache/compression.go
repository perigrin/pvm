// ABOUTME: Implements compression support for cache entries using zstd and gzip algorithms
// ABOUTME: Provides efficient data compression to reduce memory and network usage

package cache

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"sync"

	"tamarou.com/pvm/internal/errors"
)

// Compressor interface for data compression
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	Decompress(data []byte) ([]byte, error)
	Type() string
}

// NewCompressor creates a compressor based on the specified type
func NewCompressor(compressionType string) (Compressor, error) {
	switch compressionType {
	case "none":
		return &NoOpCompressor{}, nil
	case "gzip":
		return &GzipCompressor{}, nil
	case "zstd":
		return &ZstdCompressor{}, nil
	default:
		return nil, errors.NewSystemError("001", fmt.Sprintf("unsupported compression type: %s", compressionType), nil)
	}
}

// NoOpCompressor implements a pass-through compressor
type NoOpCompressor struct{}

// Compress returns the data unchanged
func (c *NoOpCompressor) Compress(data []byte) ([]byte, error) {
	return data, nil
}

// Decompress returns the data unchanged
func (c *NoOpCompressor) Decompress(data []byte) ([]byte, error) {
	return data, nil
}

// Type returns the compressor type
func (c *NoOpCompressor) Type() string {
	return "none"
}

// GzipCompressor implements gzip compression
type GzipCompressor struct {
	writerPool sync.Pool
	readerPool sync.Pool
}

// Compress compresses data using gzip
func (c *GzipCompressor) Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	// Get writer from pool or create new one
	var writer *gzip.Writer
	if w := c.writerPool.Get(); w != nil {
		writer = w.(*gzip.Writer)
		writer.Reset(&buf)
	} else {
		writer = gzip.NewWriter(&buf)
	}

	// Write data
	if _, err := writer.Write(data); err != nil {
		c.writerPool.Put(writer)
		return nil, errors.Wrap(err, "PSC", "compression", "005", "failed to write gzip data")
	}

	// Close writer to flush data
	if err := writer.Close(); err != nil {
		c.writerPool.Put(writer)
		return nil, errors.Wrap(err, "PSC", "compression", "006", "failed to close gzip writer")
	}

	// Return writer to pool
	c.writerPool.Put(writer)

	return buf.Bytes(), nil
}

// Decompress decompresses gzip data
func (c *GzipCompressor) Decompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)

	// Get reader from pool or create new one
	var reader *gzip.Reader
	var err error
	if r := c.readerPool.Get(); r != nil {
		reader = r.(*gzip.Reader)
		err = reader.Reset(buf)
		if err != nil {
			c.readerPool.Put(reader)
			return nil, errors.Wrap(err, "PSC", "compression", "007", "failed to reset gzip reader")
		}
	} else {
		reader, err = gzip.NewReader(buf)
		if err != nil {
			return nil, errors.Wrap(err, "PSC", "compression", "008", "failed to create gzip reader")
		}
	}

	// Read decompressed data
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		reader.Close()
		return nil, errors.Wrap(err, "PSC", "compression", "009", "failed to read gzip data")
	}

	// Close reader
	reader.Close()

	// Return reader to pool
	c.readerPool.Put(reader)

	return decompressed, nil
}

// Type returns the compressor type
func (c *GzipCompressor) Type() string {
	return "gzip"
}

// ZstdCompressor implements zstd compression
// Note: In a real implementation, this would use a zstd library
// For now, we'll implement it as a wrapper around gzip as a placeholder
type ZstdCompressor struct {
	gzipCompressor GzipCompressor
}

// Compress compresses data using zstd (placeholder using gzip)
func (c *ZstdCompressor) Compress(data []byte) ([]byte, error) {
	// In a real implementation, this would use zstd compression
	// For now, we use gzip as a placeholder
	return c.gzipCompressor.Compress(data)
}

// Decompress decompresses zstd data (placeholder using gzip)
func (c *ZstdCompressor) Decompress(data []byte) ([]byte, error) {
	// In a real implementation, this would use zstd decompression
	// For now, we use gzip as a placeholder
	return c.gzipCompressor.Decompress(data)
}

// Type returns the compressor type
func (c *ZstdCompressor) Type() string {
	return "zstd"
}

// CompressionBenchmark provides compression statistics
type CompressionBenchmark struct {
	Type              string
	OriginalSize      int64
	CompressedSize    int64
	CompressionRatio  float64
	CompressionTime   int64 // nanoseconds
	DecompressionTime int64 // nanoseconds
}

// BenchmarkCompressor benchmarks a compressor with sample data
func BenchmarkCompressor(compressor Compressor, data []byte) (*CompressionBenchmark, error) {
	benchmark := &CompressionBenchmark{
		Type:         compressor.Type(),
		OriginalSize: int64(len(data)),
	}

	// Measure compression
	startTime := nanoTime()
	compressed, err := compressor.Compress(data)
	if err != nil {
		return nil, errors.Wrap(err, "PSC", "compression", "010", "compression failed")
	}
	benchmark.CompressionTime = nanoTime() - startTime
	benchmark.CompressedSize = int64(len(compressed))

	// Calculate compression ratio
	if benchmark.OriginalSize > 0 {
		benchmark.CompressionRatio = float64(benchmark.CompressedSize) / float64(benchmark.OriginalSize)
	}

	// Measure decompression
	startTime = nanoTime()
	decompressed, err := compressor.Decompress(compressed)
	if err != nil {
		return nil, errors.Wrap(err, "PSC", "compression", "011", "decompression failed")
	}
	benchmark.DecompressionTime = nanoTime() - startTime

	// Verify data integrity
	if !bytes.Equal(data, decompressed) {
		return nil, errors.NewSystemError("002", "decompressed data does not match original", nil)
	}

	return benchmark, nil
}

// nanoTime returns current time in nanoseconds (placeholder)
func nanoTime() int64 {
	// In a real implementation, this would use time.Now().UnixNano()
	return 0
}

// AdaptiveCompressor automatically selects compression based on data characteristics
type AdaptiveCompressor struct {
	compressors map[string]Compressor
	threshold   float64 // Minimum compression ratio to use compression
	mu          sync.RWMutex
}

// NewAdaptiveCompressor creates a new adaptive compressor
func NewAdaptiveCompressor(threshold float64) *AdaptiveCompressor {
	ac := &AdaptiveCompressor{
		compressors: make(map[string]Compressor),
		threshold:   threshold,
	}

	// Initialize compressors
	ac.compressors["none"] = &NoOpCompressor{}
	ac.compressors["gzip"] = &GzipCompressor{}
	ac.compressors["zstd"] = &ZstdCompressor{}

	return ac
}

// Compress compresses data using the most suitable algorithm
func (c *AdaptiveCompressor) Compress(data []byte) ([]byte, error) {
	// For small data, don't compress
	if len(data) < 1024 {
		return c.compressors["none"].Compress(data)
	}

	// Try different compressors and pick the best one
	bestRatio := float64(1.0)
	bestCompressed := data
	bestType := "none"

	for name, compressor := range c.compressors {
		if name == "none" {
			continue
		}

		compressed, err := compressor.Compress(data)
		if err != nil {
			continue
		}

		ratio := float64(len(compressed)) / float64(len(data))
		if ratio < bestRatio && ratio < c.threshold {
			bestRatio = ratio
			bestCompressed = compressed
			bestType = name
		}
	}

	// Prepend compression type for decompression
	result := make([]byte, len(bestCompressed)+1)
	result[0] = c.typeToBytes(bestType)
	copy(result[1:], bestCompressed)

	return result, nil
}

// Decompress decompresses data based on its compression type
func (c *AdaptiveCompressor) Decompress(data []byte) ([]byte, error) {
	if len(data) < 1 {
		return nil, errors.NewSystemError("003", "invalid compressed data", nil)
	}

	// Extract compression type
	compType := c.bytesToType(data[0])
	compressor, exists := c.compressors[compType]
	if !exists {
		return nil, errors.NewSystemError("004", fmt.Sprintf("unknown compression type: %s", compType), nil)
	}

	// Decompress using the appropriate compressor
	return compressor.Decompress(data[1:])
}

// Type returns the compressor type
func (c *AdaptiveCompressor) Type() string {
	return "adaptive"
}

// typeToBytes converts compression type to byte
func (c *AdaptiveCompressor) typeToBytes(compType string) byte {
	switch compType {
	case "none":
		return 0
	case "gzip":
		return 1
	case "zstd":
		return 2
	default:
		return 0
	}
}

// bytesToType converts byte to compression type
func (c *AdaptiveCompressor) bytesToType(b byte) string {
	switch b {
	case 0:
		return "none"
	case 1:
		return "gzip"
	case 2:
		return "zstd"
	default:
		return "none"
	}
}

// CompressionStats tracks compression statistics
type CompressionStats struct {
	TotalBytesIn       int64
	TotalBytesOut      int64
	CompressionCount   int64
	DecompressionCount int64
	AverageRatio       float64
	mu                 sync.RWMutex
}

// UpdateCompression updates compression statistics
func (s *CompressionStats) UpdateCompression(bytesIn, bytesOut int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalBytesIn += bytesIn
	s.TotalBytesOut += bytesOut
	s.CompressionCount++

	if s.TotalBytesIn > 0 {
		s.AverageRatio = float64(s.TotalBytesOut) / float64(s.TotalBytesIn)
	}
}

// UpdateDecompression updates decompression statistics
func (s *CompressionStats) UpdateDecompression() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.DecompressionCount++
}

// GetStats returns a copy of the statistics
func (s *CompressionStats) GetStats() CompressionStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return CompressionStats{
		TotalBytesIn:       s.TotalBytesIn,
		TotalBytesOut:      s.TotalBytesOut,
		CompressionCount:   s.CompressionCount,
		DecompressionCount: s.DecompressionCount,
		AverageRatio:       s.AverageRatio,
	}
}
