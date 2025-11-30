package sqlite

import (
	"encoding/json"
	"os"
)

// EcfrAgencyRef represents a CFR reference from ecfr_agencies.json
type EcfrAgencyRef struct {
	Title   int    `json:"title"`
	Chapter string `json:"chapter"`
}

// EcfrAgency represents an agency from ecfr_agencies.json
type EcfrAgency struct {
	Name         string          `json:"name"`
	ShortName    string          `json:"short_name"`
	DisplayName  string          `json:"display_name"`
	SortableName string          `json:"sortable_name"`
	Slug         string          `json:"slug"`
	Children     []EcfrAgency    `json:"children"`
	CFRRefs      []EcfrAgencyRef `json:"cfr_references"`
}

// EcfrAgenciesRoot is the root structure of ecfr_agencies.json
type EcfrAgenciesRoot struct {
	Agencies []EcfrAgency `json:"agencies"`
}

// IngestAgencies loads agency data from a JSON file and populates the database.
// This clears and reloads all agency data on every call.
func (r *Repo) IngestAgencies(jsonPath string) error {
	// Read and parse JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return err
	}

	var root EcfrAgenciesRoot
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}

	// Clear existing data (every ETL run refreshes agency data)
	// Delete references first due to FK constraint
	if _, err := r.db.Exec("DELETE FROM agency_cfr_references"); err != nil {
		return err
	}
	if _, err := r.db.Exec("DELETE FROM agencies"); err != nil {
		return err
	}

	// Recursively insert agencies and children
	for _, agency := range root.Agencies {
		if err := r.insertAgencyTree(agency, nil); err != nil {
			return err
		}
	}

	return nil
}

// insertAgencyTree recursively inserts an agency and its children
func (r *Repo) insertAgencyTree(a EcfrAgency, parentID *string) error {
	// Use DisplayName if available, otherwise fall back to Name
	name := a.DisplayName
	if name == "" {
		name = a.Name
	}

	// Insert agency row
	_, err := r.db.Exec(`
		INSERT INTO agencies (id, name, short_name, sortable_name, parent_id)
		VALUES (?, ?, ?, ?, ?)`,
		a.Slug, name, a.ShortName, a.SortableName, parentID)
	if err != nil {
		return err
	}

	// Insert all CFR references (no ON CONFLICT - allows duplicates for N:N mapping)
	for _, ref := range a.CFRRefs {
		_, err := r.db.Exec(`
			INSERT INTO agency_cfr_references (agency_id, title, chapter)
			VALUES (?, ?, ?)`,
			a.Slug, ref.Title, ref.Chapter)
		if err != nil {
			return err
		}
	}

	// Recursively insert children with this agency as parent
	for _, child := range a.Children {
		if err := r.insertAgencyTree(child, &a.Slug); err != nil {
			return err
		}
	}

	return nil
}
