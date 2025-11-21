package govinfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTitleXML_AgencyID(t *testing.T) {
	// Create a temporary XML file simulating the structure
	xmlContent := `<?xml version="1.0" encoding="UTF-8" ?>
<DLPSTEXTCLASS>
<HEADER></HEADER>
<TEXT>
<BODY>
<DIV1 N="1" TYPE="TITLE">
	<DIV3 N="I" TYPE="CHAPTER">
		<HEAD>CHAPTER I—TEST AGENCY</HEAD>
		<DIV8 N="§ 1.1" TYPE="SECTION">
			<HEAD>§ 1.1 Test Section.</HEAD>
			<P>Content of section 1.1</P>
		</DIV8>
	</DIV3>
	<DIV3 N="II" TYPE="CHAPTER">
		<HEAD>CHAPTER II—ANOTHER AGENCY</HEAD>
		<DIV8 N="§ 2.1" TYPE="SECTION">
			<HEAD>§ 2.1 Test Section.</HEAD>
			<P>Content of section 2.1</P>
		</DIV8>
	</DIV3>
</DIV1>
</BODY>
</TEXT>
</DLPSTEXTCLASS>`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.xml")
	if err := os.WriteFile(tmpFile, []byte(xmlContent), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	client := &Client{} // We don't need a full client for this test
	sections, err := client.ParseTitleXML(tmpFile)
	if err != nil {
		t.Fatalf("ParseTitleXML failed: %v", err)
	}

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	// Check first section
	if sections[0].AgencyID != "I" {
		t.Errorf("Section 1: Expected AgencyID 'I', got '%s'", sections[0].AgencyID)
	}
	if sections[0].ID != "§ 1.1" {
		t.Errorf("Section 1: Expected ID '§ 1.1', got '%s'", sections[0].ID)
	}

	// Check second section
	if sections[1].AgencyID != "II" {
		t.Errorf("Section 2: Expected AgencyID 'II', got '%s'", sections[1].AgencyID)
	}
	if sections[1].ID != "§ 2.1" {
		t.Errorf("Section 2: Expected ID '§ 2.1', got '%s'", sections[1].ID)
	}
}
