package index

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const snapshotVersion = 1

var (
	ErrNilIndexForSnapshot        = errors.New("index is nil")
	ErrInvalidIndexState          = errors.New("index contains nil required fields")
	ErrNegativeIndexStats         = errors.New("index contains negative stats")
	ErrInvalidNextDocID           = errors.New("index has invalid nextDocId")
	ErrNilSnapshot                = errors.New("indexSnapshot is nil")
	ErrUnsupportedSnapshotVersion = errors.New("incorrect version ,mismatch")
	ErrInvalidSnapshotState       = errors.New("indexSnapshot's state is invalid")
	ErrNegativeSnapshotStats      = errors.New("stats are awlays non-negative :error")
	ErrInvalidSnapshotNextDocID   = errors.New("invalid docid doesnt exist")

	ErrEmptyPath = errors.New("path is empty")
)

type indexSnapshot struct {
	Version       int                    `json:"version"`
	DocTable      map[int]Document       `json:"doc_table"`
	URLDedup      map[string]int         `json:"url_dedup"`
	InvertedIndex map[string]map[int]int `json:"inverted_index"`
	DocLen        map[int]int            `json:"doc_len"`
	TotalDocs     int                    `json:"total_docs"`
	TotalDocLen   int                    `json:"total_doc_len"`
	AvgDocLen     float64                `json:"avg_doc_len"`
	NextDocID     int                    `json:"next_doc_id"`
	DocFreq       map[string]int         `json:"doc_freq"`
}

func (idx *Index) toSnapshot() (*indexSnapshot, error) {
	if idx == nil {
		return nil, ErrNilIndexForSnapshot
	}

	if idx.docTable == nil ||
		idx.invertedIndex == nil ||
		idx.docLen == nil ||
		idx.docFreq == nil ||
		idx.urlDedup == nil {
		return nil, ErrInvalidIndexState
	}

	if idx.totalDocs < 0 || idx.totalDocLen < 0 || idx.avgDocLen < 0 {
		return nil, ErrNegativeIndexStats
	}

	if idx.nextDocId < 1 {
		return nil, ErrInvalidNextDocID
	}

	snapshot := &indexSnapshot{
		Version:       snapshotVersion,
		DocTable:      idx.docTable,
		URLDedup:      idx.urlDedup,
		InvertedIndex: idx.invertedIndex,
		DocLen:        idx.docLen,
		TotalDocs:     idx.totalDocs,
		TotalDocLen:   idx.totalDocLen,
		AvgDocLen:     idx.avgDocLen,
		NextDocID:     idx.nextDocId,
		DocFreq:       idx.docFreq,
	}

	return snapshot, nil
}

func fromSnapshot(s *indexSnapshot) (*Index, error) {
	if s == nil {
		return nil, ErrNilSnapshot
	}

	if s.DocTable == nil ||
		s.URLDedup == nil ||
		s.InvertedIndex == nil ||
		s.DocLen == nil ||
		s.DocFreq == nil {
		return nil, ErrInvalidSnapshotState
	}

	if s.TotalDocs < 0 || s.TotalDocLen < 0 || s.AvgDocLen < 0 {
		return nil, ErrNegativeSnapshotStats
	}

	if s.NextDocID < 1 {
		return nil, ErrInvalidSnapshotNextDocID
	}

	// Defensive recompute of avgDocLen from totals.
	avgDocLen := 0.0
	if s.TotalDocs > 0 {
		avgDocLen = float64(s.TotalDocLen) / float64(s.TotalDocs)
	}

	idx := &Index{
		docTable:      s.DocTable,
		urlDedup:      s.URLDedup,
		invertedIndex: s.InvertedIndex,
		docLen:        s.DocLen,

		totalDocs:   s.TotalDocs,
		totalDocLen: s.TotalDocLen,
		avgDocLen:   avgDocLen,
		nextDocId:   s.NextDocID,
		docFreq:     s.DocFreq,
	}

	return idx, nil
}
func (idx *Index) Save(path string) error {
	if path == "" {
		return ErrEmptyPath
	}

	snapshot, err := idx.toSnapshot()
	if err != nil {
		return fmt.Errorf("toSnapshot failed: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot failed: %w", err)
	}

	if err := ensureDir(path); err != nil {
		return fmt.Errorf("ensure directory failed: %w", err)
	}

	if err := writeFileAtomic(path, data); err != nil {
		return fmt.Errorf("atomic write failed: %w", err)
	}

	return nil
}

func Load(path string) (*Index, error) {
	if path == "" {
		return nil, ErrEmptyPath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot file failed: %w", err)
	}

	var snap indexSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot failed: %w", err)
	}

	idx, err := fromSnapshot(&snap)
	if err != nil {
		return nil, fmt.Errorf("fromSnapshot failed: %w", err)
	}

	return idx, nil
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "snapshot-*.tmp")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()

	// cleanup temp file on any failure path
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return err
	}

	if err := tmp.Sync(); err != nil {
		cleanup()
		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	// atomic replace (same filesystem)
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return nil
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755) //owner group and others permissibility
}
