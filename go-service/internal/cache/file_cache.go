package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wbentaleb/student-report-service/internal/dto"
)

type FileCache struct {
	ttl      time.Duration
	basePath string
	data     map[string]CacheEntry // In-memory data (id:hash → CacheEntry with file path)
	mu       sync.RWMutex
}

type CacheEntry struct {
	FilePath  string
	ExpiresAt time.Time
}

func NewFileCache(basePath string, ttl time.Duration) (*FileCache, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Clean up any existing PDF files from previous runs
	if err := cleanupDirectory(basePath); err != nil {
		return nil, fmt.Errorf("failed to cleanup cache directory: %w", err)
	}

	fc := &FileCache{
		data:     make(map[string]CacheEntry),
		basePath: basePath,
		ttl:      ttl,
	}

	fc.startCleanupWorker()

	return fc, nil
}

// cleanupDirectory removes all PDF files in the cache directory
func cleanupDirectory(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dirPath, entry.Name())
		if err := os.Remove(filePath); err != nil {
			continue
		}
	}

	return nil
}

// GenerateStudentHash GenerateHash creates a hash from student data for cache key
func GenerateStudentHash(student *dto.Student) string {
	data := fmt.Sprintf("%v:%v:%v:%v:%v",
		student.Name,
		student.Class,
		student.Section,
		student.LastUpdated,
		student.AdmissionDate,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

func (c *FileCache) Get(studentID, hash string) ([]byte, bool) {
	c.mu.RLock()
	key := fmt.Sprintf("%s:%s", studentID, hash)
	entry, exists := c.data[key]
	c.mu.RUnlock()

	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	// Read file from disk
	data, err := os.ReadFile(entry.FilePath)
	if err != nil {
		// File missing, clean up index
		c.mu.Lock()
		delete(c.data, key)
		c.mu.Unlock()
		return nil, false
	}

	return data, true
}

func (c *FileCache) Set(studentID string, data []byte, hash string) error {
	key := fmt.Sprintf("%s:%s", studentID, hash)
	filename := fmt.Sprintf("student_%s_%s.pdf", studentID, hash)
	filePath := filepath.Join(c.basePath, filename)

	// Write file to disk
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Update index
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean old versions of same student
	for k, entry := range c.data {
		// Parse studentID from the key (format: "studentID:hash")
		if strings.HasPrefix(k, studentID+":") && k != key {
			os.Remove(entry.FilePath) // Clean old file
			delete(c.data, k)
		}
	}

	c.data[key] = CacheEntry{
		FilePath:  filePath,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	return nil
}

// Cleanup worker runs periodically to remove expired entries
func (c *FileCache) startCleanupWorker() {
	ticker := time.NewTicker(time.Minute)

	go func() {
		for range ticker.C {
			c.removeExpiredFiles()
		}
	}()
}

func (c *FileCache) removeExpiredFiles() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			os.Remove(entry.FilePath) // Remove file from disk
			delete(c.data, key)       // Remove from index
		}
	}
}
