package cache

type PDFCache interface {
	Get(studentID, hash string) ([]byte, bool)
	Set(studentID string, data []byte, hash string) error
}
