/* parse_test.go provides unit tests for the .DNA file format parser. */
package dnafile

import (
	"testing"
)

const exampleFilePath = "./fixtures/example.dna"

// TestParseHeader tests the ParseHeader function.
func TestParseSequenceAndSequenceProperties(t *testing.T) {

	record, err := NewDnaFileReader(exampleFilePath).Parse()
	if err != nil {
		t.Errorf("Error parsing file: %v", err)
		return
	}

	if record.FilePath != exampleFilePath {
		t.Errorf("Expected FilePath to be %s, got %s", exampleFilePath, record.FilePath)
	}

	if record.Sequence[0:10] != "ATCCGGATAT" {
		t.Errorf("Expected Sequence to be ATCCGGATAT, got %s", record.Sequence[0:10])
		return
	}

	if record.Sequence[len(record.Sequence)-10:] != "CCATTCGCCA" {
		t.Errorf("Expected Sequence to end with CCATTCGCCA, got %s", record.Sequence[len(record.Sequence)-10:])
		return
	}

	if record.SequenceProperties.Topology != "circular" {
		t.Errorf("Expected Topology to be circular, got %s", record.SequenceProperties.Topology)
		return
	}

	if record.SequenceProperties.Stranded != "double" {
		t.Errorf("Expected Stranded to be double, got %s", record.SequenceProperties.Stranded)
		return
	}

	if record.SequenceProperties.Length != 5493 {
		t.Errorf("Expected Length to be 5493, got %d", record.SequenceProperties.Length)
		return
	}

	if record.SequenceProperties.AMethylated != true {
		t.Errorf("Expected AMethylated to be true, got %t", record.SequenceProperties.AMethylated)
		return
	}
}

func TestParseMeta(t *testing.T) {
	record, err := NewDnaFileReader(exampleFilePath).Parse()
	if err != nil {
		t.Errorf("Error parsing file: %v", err)
		return
	}

	if record.Meta.IsDna != true {
		t.Errorf("Expected is_dna to be true, got %t", record.Meta.IsDna)
		return
	}

	if record.Meta.ExportVersion != 15 {
		t.Errorf("Expected ExportVersion to be 15, got %d", record.Meta.ExportVersion)
		return
	}

	if record.Meta.ImportVersion != 19 {
		t.Errorf("Expected ImportVersion to be 19, got %d", record.Meta.ImportVersion)
		return
	}

}

// Test that the notes of the file are returned correctly.
func TestParseNotes(t *testing.T) {
	record, err := NewDnaFileReader(exampleFilePath).Parse()
	if err != nil {
		t.Errorf("Error parsing file: %v", err)
		return
	}

	if len(record.Notes) != 13 {
		t.Errorf("Expected 13 notes, got %d", len(record.Notes))
		return
	}

	expectedNotes := map[string]string{
		"SequenceClass":           "UNA",
		"TransformedInto":         "Unspecified",
		"Type":                    "Synthetic",
		"LastModified":            "2024.3.6",
		"Organism":                "Escherichia coli",
		"UseCustomMapLabel":       "1",
		"Description":             "<html><body>Bacterial vector that encodes a signal sequence for inducible expression of proteins in the periplasm.</body></html>",
		"Created":                 "2012.5.12",
		"CreatedBy":               "MilliporeSigma (Novagen)",
		"Comments":                "<html><body><br></body></html>",
		"UUID":                    "4b7eca01-cceb-44ba-b754-6efcad9cc573",
		"ConfirmedExperimentally": "0",
		"CustomMapLabel":          "pET-22b(+)",
	}

	for key, value := range record.Notes {
		expectedVal, ok := expectedNotes[key]
		if ok {
			if value != expectedVal {
				t.Errorf("Expected %s to be %s, got %s", key, expectedVal, value)
			}
		} else {
			t.Errorf("Unexpected note: %s", key)
		}
	}
}

func TestParseFeatures(t *testing.T) {
	record, err := NewDnaFileReader(exampleFilePath).Parse()
	if err != nil {
		t.Errorf("Error parsing file: %v", err)
		return
	}

	if len(record.Features) != 14 {
		t.Errorf("Expected 14 feature, got %d", len(record.Features))
		return
	}

	firstFeature := record.Features[0]
	if firstFeature.Type != "RBS" {
		t.Errorf("Expected first feature to be of type RBS, got %s", firstFeature.Type)
	}

	if firstFeature.Start != 298 {
		t.Errorf("Expected first feature to start at 298, got %d", firstFeature.Start)
	}

	if firstFeature.End != 303 {
		t.Errorf("Expected first feature to end at 303, got %d", firstFeature.End)
	}

	if len(firstFeature.Segments) != 1 {
		t.Errorf("Expected first feature to have 5 segments, got %d", len(firstFeature.Segments))
	}

	if firstFeature.Segments[0].Start != 298 {
		t.Errorf("Expected first feature segment to start at 298, got %d", firstFeature.Segments[0].Start)
	}

	if firstFeature.Segments[0].End != 303 {
		t.Errorf("Expected first feature segment to end at 303, got %d", firstFeature.Segments[0].End)
	}

	if firstFeature.Segments[0].Color != "#a6acb3" {
		t.Errorf("Expected first feature segment color to be #a6acb3, got %s", firstFeature.Segments[0].Color)
	}

	if firstFeature.Segments[0].IsTranslated != false {
		t.Errorf("Expected first feature segment to be isTranslated false, got %v", firstFeature.Segments[0].IsTranslated)
	}
}
