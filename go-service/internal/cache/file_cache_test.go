package cache

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wbentaleb/student-report-service/internal/dto"
)

func TestNewFileCache_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour

	// Execute
	cache, err := NewFileCache(tempDir, ttl)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, tempDir, cache.basePath)
	assert.Equal(t, ttl, cache.ttl)
	assert.NotNil(t, cache.data)
}

func TestNewFileCache_CreateDirectory(t *testing.T) {
	// Setup
	tempDir := filepath.Join(t.TempDir(), "new_cache_dir")
	ttl := 1 * time.Hour

	// Ensure directory doesn't exist
	_, err := os.Stat(tempDir)
	require.True(t, os.IsNotExist(err))

	// Execute
	cache, err := NewFileCache(tempDir, ttl)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, cache)

	// Verify directory was created
	info, err := os.Stat(tempDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNewFileCache_CleanupExistingFiles(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour

	// Create some existing files
	testFile1 := filepath.Join(tempDir, "old_file1.pdf")
	testFile2 := filepath.Join(tempDir, "old_file2.pdf")
	require.NoError(t, os.WriteFile(testFile1, []byte("old content 1"), 0644))
	require.NoError(t, os.WriteFile(testFile2, []byte("old content 2"), 0644))

	// Verify files exist
	_, err := os.Stat(testFile1)
	require.NoError(t, err)
	_, err = os.Stat(testFile2)
	require.NoError(t, err)

	// Execute
	cache, err := NewFileCache(tempDir, ttl)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, cache)

	// Verify old files were cleaned up
	_, err = os.Stat(testFile1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(testFile2)
	assert.True(t, os.IsNotExist(err))
}

func TestGenerateStudentHash(t *testing.T) {
	// Setup
	student := &dto.Student{
		Name:          "John Doe",
		Class:         "10",
		Section:       "A",
		LastUpdated:   "2024-01-01T10:00:00Z",
		AdmissionDate: "2015-06-01T00:00:00Z",
	}

	// Execute
	hash1 := GenerateStudentHash(student)

	// Assert
	assert.NotEmpty(t, hash1)
	assert.Len(t, hash1, 16) // Should be 16 characters

	// Test consistency - same data should produce same hash
	hash2 := GenerateStudentHash(student)
	assert.Equal(t, hash1, hash2)

	// Test uniqueness - different data should produce different hash
	student.Name = "Jane Doe"
	hash3 := GenerateStudentHash(student)
	assert.NotEqual(t, hash1, hash3)
}

func TestGenerateStudentHash_EmptyStudent(t *testing.T) {
	// Setup
	student := &dto.Student{}

	// Execute
	hash := GenerateStudentHash(student)

	// Assert
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 16)
}

func TestFileCache_SetAndGet_Success(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	studentID := "12345"
	hash := "abcd1234"
	pdfData := []byte("test pdf content")

	// Execute Set
	err = cache.Set(studentID, pdfData, hash)
	require.NoError(t, err)

	// Execute Get
	retrievedData, found := cache.Get(studentID, hash)

	// Assert
	assert.True(t, found)
	assert.Equal(t, pdfData, retrievedData)

	// Verify file exists on disk
	expectedFile := filepath.Join(tempDir, "student_12345_abcd1234.pdf")
	_, err = os.Stat(expectedFile)
	assert.NoError(t, err)
}

func TestFileCache_Get_NotFound(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	// Execute
	data, found := cache.Get("nonexistent", "hash123")

	// Assert
	assert.False(t, found)
	assert.Nil(t, data)
}

func TestFileCache_Get_Expired(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 100 * time.Millisecond // Short TTL for testing
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	studentID := "12345"
	hash := "abcd1234"
	pdfData := []byte("test pdf content")

	// Set data
	err = cache.Set(studentID, pdfData, hash)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Execute
	data, found := cache.Get(studentID, hash)

	// Assert
	assert.False(t, found)
	assert.Nil(t, data)
}

func TestFileCache_Get_FileMissing(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	studentID := "12345"
	hash := "abcd1234"
	pdfData := []byte("test pdf content")

	// Set data
	err = cache.Set(studentID, pdfData, hash)
	require.NoError(t, err)

	// Manually delete the file
	expectedFile := filepath.Join(tempDir, "student_12345_abcd1234.pdf")
	err = os.Remove(expectedFile)
	require.NoError(t, err)

	// Execute
	data, found := cache.Get(studentID, hash)

	// Assert
	assert.False(t, found)
	assert.Nil(t, data)

	// Verify entry was cleaned from index
	cache.mu.RLock()
	_, exists := cache.data["12345:abcd1234"]
	cache.mu.RUnlock()
	assert.False(t, exists)
}

func TestFileCache_Set_OverwriteOldVersion(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	studentID := "12345"
	oldHash := "oldhash"
	newHash := "newhash"
	oldData := []byte("old pdf content")
	newData := []byte("new pdf content")

	// Set old version
	err = cache.Set(studentID, oldData, oldHash)
	require.NoError(t, err)

	oldFile := filepath.Join(tempDir, "student_12345_oldhash.pdf")
	_, err = os.Stat(oldFile)
	require.NoError(t, err)

	// Set new version
	err = cache.Set(studentID, newData, newHash)
	require.NoError(t, err)

	// Assert old version is removed
	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err))

	// Assert new version exists
	newFile := filepath.Join(tempDir, "student_12345_newhash.pdf")
	_, err = os.Stat(newFile)
	assert.NoError(t, err)

	// Verify old version not in cache
	data, found := cache.Get(studentID, oldHash)
	assert.False(t, found)
	assert.Nil(t, data)

	// Verify new version is in cache
	data, found = cache.Get(studentID, newHash)
	assert.True(t, found)
	assert.Equal(t, newData, data)
}

func TestFileCache_Set_MultipleStudents(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	// Set data for multiple students
	testCases := []struct {
		studentID string
		hash      string
		data      []byte
	}{
		{"111", "hash111", []byte("student 111 pdf")},
		{"222", "hash222", []byte("student 222 pdf")},
		{"333", "hash333", []byte("student 333 pdf")},
	}

	// Execute Set for all
	for _, tc := range testCases {
		err := cache.Set(tc.studentID, tc.data, tc.hash)
		require.NoError(t, err)
	}

	// Verify all can be retrieved
	for _, tc := range testCases {
		data, found := cache.Get(tc.studentID, tc.hash)
		assert.True(t, found)
		assert.Equal(t, tc.data, data)
	}
}

func TestFileCache_RemoveExpiredFiles(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 100 * time.Millisecond
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	// Add entries
	cache.Set("student1", []byte("data1"), "hash1")
	cache.Set("student2", []byte("data2"), "hash2")

	// Verify files exist
	file1 := filepath.Join(tempDir, "student_student1_hash1.pdf")
	file2 := filepath.Join(tempDir, "student_student2_hash2.pdf")
	_, err = os.Stat(file1)
	require.NoError(t, err)
	_, err = os.Stat(file2)
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Manually trigger cleanup
	cache.removeExpiredFiles()

	// Verify files are removed
	_, err = os.Stat(file1)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(file2)
	assert.True(t, os.IsNotExist(err))

	// Verify index is cleaned
	cache.mu.RLock()
	assert.Len(t, cache.data, 0)
	cache.mu.RUnlock()
}

func TestFileCache_ConcurrentAccess(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	// Concurrent operations
	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			studentID := string(rune('0' + id))
			hash := "hash" + studentID
			data := []byte("data" + studentID)
			err := cache.Set(studentID, data, hash)
			assert.NoError(t, err)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			studentID := string(rune('0' + id))
			hash := "hash" + studentID
			// Small delay to ensure writes happen first
			time.Sleep(10 * time.Millisecond)
			_, _ = cache.Get(studentID, hash)
		}(i)
	}

	wg.Wait()

	// Verify data integrity
	for i := 0; i < numGoroutines; i++ {
		studentID := string(rune('0' + i))
		hash := "hash" + studentID
		expectedData := []byte("data" + studentID)

		data, found := cache.Get(studentID, hash)
		assert.True(t, found)
		assert.Equal(t, expectedData, data)
	}
}

func TestFileCache_Set_WriteError(t *testing.T) {
	// Setup - use a directory that will cause write errors
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	// Make directory read-only to cause write error
	err = os.Chmod(tempDir, 0555)
	require.NoError(t, err)

	// Restore permissions at the end
	defer func() {
		os.Chmod(tempDir, 0755)
	}()

	// Execute
	studentID := "12345"
	hash := "abcd1234"
	pdfData := []byte("test pdf content")
	err = cache.Set(studentID, pdfData, hash)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write cache file")
}

func TestFileCache_CleanupWorker(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 100 * time.Millisecond
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	// Add data
	cache.Set("student1", []byte("data1"), "hash1")

	file := filepath.Join(tempDir, "student_student1_hash1.pdf")
	_, err = os.Stat(file)
	require.NoError(t, err)

	// Wait for automatic cleanup (cleanup runs every minute, but our TTL is 100ms)
	// For testing, we'll manually trigger instead of waiting
	time.Sleep(150 * time.Millisecond)
	cache.removeExpiredFiles()

	// Verify file is removed
	_, err = os.Stat(file)
	assert.True(t, os.IsNotExist(err))
}

func TestCleanupDirectory_SkipsDirectories(t *testing.T) {
	// Setup
	tempDir := t.TempDir()

	// Create a subdirectory and a file
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(tempDir, "test.pdf")
	err = os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Execute
	err = cleanupDirectory(tempDir)
	require.NoError(t, err)

	// Assert subdirectory still exists
	_, err = os.Stat(subDir)
	assert.NoError(t, err)

	// Assert file was removed
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
}

func TestFileCache_LargePDFData(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	ttl := 1 * time.Hour
	cache, err := NewFileCache(tempDir, ttl)
	require.NoError(t, err)

	// Create large PDF data (5MB)
	largePDFData := make([]byte, 5*1024*1024)
	for i := range largePDFData {
		largePDFData[i] = byte(i % 256)
	}

	studentID := "12345"
	hash := "largehash"

	// Execute
	err = cache.Set(studentID, largePDFData, hash)
	require.NoError(t, err)

	retrievedData, found := cache.Get(studentID, hash)

	// Assert
	assert.True(t, found)
	assert.Equal(t, largePDFData, retrievedData)
}

// Benchmark tests
func BenchmarkGenerateStudentHash(b *testing.B) {
	student := &dto.Student{
		Name:          "John Doe",
		Class:         "10",
		Section:       "A",
		LastUpdated:   "2024-01-01T10:00:00Z",
		AdmissionDate: "2015-06-01T00:00:00Z",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateStudentHash(student)
	}
}

func BenchmarkFileCache_Set(b *testing.B) {
	tempDir := b.TempDir()
	cache, _ := NewFileCache(tempDir, 1*time.Hour)

	pdfData := []byte("benchmark pdf content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		studentID := string(rune(i))
		hash := "hash" + studentID
		_ = cache.Set(studentID, pdfData, hash)
	}
}

func BenchmarkFileCache_Get(b *testing.B) {
	tempDir := b.TempDir()
	cache, _ := NewFileCache(tempDir, 1*time.Hour)

	// Pre-populate cache
	studentID := "12345"
	hash := "benchhash"
	pdfData := []byte("benchmark pdf content")
	cache.Set(studentID, pdfData, hash)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get(studentID, hash)
	}
}
