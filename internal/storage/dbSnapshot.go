package storage

import (
	"database/sql"
	"fmt"

	"github.com/waiyneee/Simplesearch/internal/index"
)

func SaveIndex(db *sql.DB, idx *index.Index) error {
	if db == nil || idx == nil {
		return fmt.Errorf("db or index is nil")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM terms`); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM documents`); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM stats`); err != nil {
		return err
	}

	// Insert documents
	docStmt, err := tx.Prepare(`
INSERT INTO documents (id, url, title, body, length)
VALUES (?, ?, ?, ?, ?)
`)
	if err != nil {
		return err
	}
	defer docStmt.Close()

	for id, doc := range idx.DocTable() {
		if _, err = docStmt.Exec(id, doc.URL, doc.Title, doc.Body, doc.Length); err != nil {
			return err
		}
	}

	// Insert terms
	termStmt, err := tx.Prepare(`
INSERT INTO terms (term, doc_id, freq)
VALUES (?, ?, ?)
`)
	if err != nil {
		return err
	}
	defer termStmt.Close()

	for term, postings := range idx.InvertedIndex() {
		for docID, tf := range postings {
			if _, err = termStmt.Exec(term, docID, tf); err != nil {
				return err
			}
		}
	}

	// Insert stats (single row)
	_, err = tx.Exec(`
INSERT INTO stats (id, total_docs, total_doc_len, next_doc_id)
VALUES (1, ?, ?, ?)
`, idx.TotalDocs(), idx.TotalDocLen(), idx.NextDocID())
	if err != nil {
		return err
	}

	return tx.Commit()
}

func LoadIndex(db *sql.DB) (*index.Index, error) {
	if db == nil {
		return nil, fmt.Errorf("db is nil")
	}

	idx := index.New()

	// Load documents
	rows, err := db.Query(`SELECT id, url, title, body, length FROM documents`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var doc index.Document
		if err := rows.Scan(&doc.ID, &doc.URL, &doc.Title, &doc.Body, &doc.Length); err != nil {
			return nil, err
		}
		idx.AddDocumentFromDB(doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load terms
	rows, err = db.Query(`SELECT term, doc_id, freq FROM terms`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var term string
		var docID, freq int
		if err := rows.Scan(&term, &docID, &freq); err != nil {
			return nil, err
		}
		idx.AddPostingFromDB(term, docID, freq)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load stats
	var totalDocs, totalDocLen, nextDocID int
	err = db.QueryRow(`SELECT total_docs, total_doc_len, next_doc_id FROM stats WHERE id = 1`).
		Scan(&totalDocs, &totalDocLen, &nextDocID)
	if err == nil {
		idx.SetStatsFromDB(totalDocs, totalDocLen, nextDocID)
	}

	return idx, nil
}
