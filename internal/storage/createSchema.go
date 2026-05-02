package storage

import "database/sql"

func CreateSchema(db *sql.DB) error {
	query := `
CREATE TABLE IF NOT EXISTS documents (
  id INTEGER PRIMARY KEY,
  url TEXT UNIQUE,
  title TEXT,
  body TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS terms (
  term TEXT NOT NULL,
  doc_id INTEGER NOT NULL,
  freq INTEGER NOT NULL,
  PRIMARY KEY (term, doc_id),
  FOREIGN KEY (doc_id) REFERENCES documents(id)
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX IF NOT EXISTS idx_terms_term ON terms(term);


CREATE INDEX IF NOT EXISTS idx_terms_doc ON terms(doc_id);
`
	_, err := db.Exec(query)
	return err
}