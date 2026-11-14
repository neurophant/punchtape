// Package throttle provides bandwidth throttling using a token bucket algorithm.
// This helps prevent traffic bursts that can trigger DDoS detection by ISPs.
package throttle

import (
	"context"
	"io"
	"sync"
	"time"
)

// Config holds throttling configuration
type Config struct {
	// MaxBytesPerSecond limits the maximum throughput in bytes per second.
	// A value of 0 means unlimited.
	MaxBytesPerSecond int64

	// BurstSize is the maximum number of bytes that can be sent in a burst.
	// Should be at least as large as the maximum expected packet size.
	// Default: 64KB if MaxBytesPerSecond is set
	BurstSize int64
}

// DefaultConfig returns a default throttling config (unlimited)
func DefaultConfig() Config {
	return Config{
		MaxBytesPerSecond: 0, // Unlimited
		BurstSize:         64 * 1024,
	}
}

// Bucket implements a token bucket rate limiter for bandwidth control.
// Tokens represent bytes that can be transmitted.
type Bucket struct {
	mu         sync.Mutex
	tokens     float64   // Current number of tokens (bytes)
	maxTokens  float64   // Maximum bucket size (burst size)
	refillRate float64   // Tokens added per nanosecond
	lastRefill time.Time // Last time tokens were refilled
	unlimited  bool      // Whether throttling is disabled
}

// NewBucket creates a new token bucket with the given configuration.
// If maxBytesPerSecond is 0 or negative, throttling is disabled.
func NewBucket(cfg Config) *Bucket {
	if cfg.MaxBytesPerSecond <= 0 {
		return &Bucket{unlimited: true}
	}

	burstSize := cfg.BurstSize
	if burstSize <= 0 {
		burstSize = 64 * 1024 // Default 64KB burst
	}

	// Ensure burst size is at least 1 second of bandwidth
	minBurst := cfg.MaxBytesPerSecond
	if burstSize < minBurst {
		burstSize = minBurst
	}

	return &Bucket{
		tokens:     float64(burstSize), // Start with full bucket
		maxTokens:  float64(burstSize),
		refillRate: float64(cfg.MaxBytesPerSecond) / float64(time.Second), // tokens per nanosecond
		lastRefill: time.Now(),
		unlimited:  false,
	}
}

// Wait blocks until n bytes can be transmitted, respecting the rate limit.
// Returns an error if the context is cancelled.
func (b *Bucket) Wait(ctx context.Context, n int) error {
	if b.unlimited {
		return nil
	}

	for {
		waitTime := b.take(n)
		if waitTime == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again after waiting
		}
	}
}

// WaitN is an alias for Wait for compatibility
func (b *Bucket) WaitN(ctx context.Context, n int) error {
	return b.Wait(ctx, n)
}

// take attempts to take n tokens from the bucket.
// Returns 0 if successful, or the duration to wait if not enough tokens.
func (b *Bucket) take(n int) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	b.tokens += float64(elapsed) * b.refillRate
	if b.tokens > b.maxTokens {
		b.tokens = b.maxTokens
	}
	b.lastRefill = now

	// Check if we have enough tokens
	needed := float64(n)
	if b.tokens >= needed {
		b.tokens -= needed
		return 0
	}

	// Calculate wait time for more tokens
	deficit := needed - b.tokens
	waitNanos := deficit / b.refillRate
	return time.Duration(waitNanos)
}

// Available returns the current number of available tokens (bytes)
func (b *Bucket) Available() int64 {
	if b.unlimited {
		return -1 // Unlimited
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	tokens := b.tokens + float64(elapsed)*b.refillRate
	if tokens > b.maxTokens {
		tokens = b.maxTokens
	}

	return int64(tokens)
}

// IsUnlimited returns true if throttling is disabled
func (b *Bucket) IsUnlimited() bool {
	return b.unlimited
}

// ThrottledReader wraps an io.Reader with bandwidth throttling
type ThrottledReader struct {
	reader io.Reader
	bucket *Bucket
	ctx    context.Context
}

// NewThrottledReader creates a new throttled reader
func NewThrottledReader(ctx context.Context, r io.Reader, bucket *Bucket) *ThrottledReader {
	return &ThrottledReader{
		reader: r,
		bucket: bucket,
		ctx:    ctx,
	}
}

// Read implements io.Reader with throttling
func (r *ThrottledReader) Read(p []byte) (int, error) {
	// Wait for tokens before reading
	if err := r.bucket.Wait(r.ctx, len(p)); err != nil {
		return 0, err
	}

	n, err := r.reader.Read(p)
	return n, err
}

// ThrottledWriter wraps an io.Writer with bandwidth throttling
type ThrottledWriter struct {
	writer io.Writer
	bucket *Bucket
	ctx    context.Context
}

// NewThrottledWriter creates a new throttled writer
func NewThrottledWriter(ctx context.Context, w io.Writer, bucket *Bucket) *ThrottledWriter {
	return &ThrottledWriter{
		writer: w,
		bucket: bucket,
		ctx:    ctx,
	}
}

// Write implements io.Writer with throttling
func (w *ThrottledWriter) Write(p []byte) (int, error) {
	// Wait for tokens before writing
	if err := w.bucket.Wait(w.ctx, len(p)); err != nil {
		return 0, err
	}

	n, err := w.writer.Write(p)
	return n, err
}

// Limiter manages bandwidth limiting for multiple streams
type Limiter struct {
	uploadBucket   *Bucket
	downloadBucket *Bucket
}

// NewLimiter creates a new bandwidth limiter
func NewLimiter(uploadBytesPerSec, downloadBytesPerSec int64) *Limiter {
	return &Limiter{
		uploadBucket: NewBucket(Config{
			MaxBytesPerSecond: uploadBytesPerSec,
			BurstSize:         64 * 1024,
		}),
		downloadBucket: NewBucket(Config{
			MaxBytesPerSecond: downloadBytesPerSec,
			BurstSize:         64 * 1024,
		}),
	}
}

// NewSymmetricLimiter creates a limiter with same upload/download limits
func NewSymmetricLimiter(bytesPerSec int64) *Limiter {
	return NewLimiter(bytesPerSec, bytesPerSec)
}

// WrapReader wraps a reader with download throttling
func (l *Limiter) WrapReader(ctx context.Context, r io.Reader) io.Reader {
	if l.downloadBucket.IsUnlimited() {
		return r
	}
	return NewThrottledReader(ctx, r, l.downloadBucket)
}

// WrapWriter wraps a writer with upload throttling
func (l *Limiter) WrapWriter(ctx context.Context, w io.Writer) io.Writer {
	if l.uploadBucket.IsUnlimited() {
		return w
	}
	return NewThrottledWriter(ctx, w, l.uploadBucket)
}

// WaitUpload waits for permission to upload n bytes
func (l *Limiter) WaitUpload(ctx context.Context, n int) error {
	return l.uploadBucket.Wait(ctx, n)
}

// WaitDownload waits for permission to download n bytes
func (l *Limiter) WaitDownload(ctx context.Context, n int) error {
	return l.downloadBucket.Wait(ctx, n)
}

// IsUnlimited returns true if both upload and download are unlimited
func (l *Limiter) IsUnlimited() bool {
	return l.uploadBucket.IsUnlimited() && l.downloadBucket.IsUnlimited()
}
