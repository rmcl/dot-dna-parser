package dnafile

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"os"
)

type DnaFileReader struct {
	filePath string
}

func NewDnaFileReader(filePath string) *DnaFileReader {
	return &DnaFileReader{
		filePath: filePath,
	}
}

// Create a new record for holding the contents of a dna file.
func newDnaFileRecord(filePath string) *DnaFileRecord {
	return &DnaFileRecord{
		FilePath: filePath,
		//Primers:       []Primer{},
		Features:           make([]Feature, 0),
		Notes:              make(map[string]string),
		SequenceProperties: SequenceProperties{},
		Meta:               Meta{},
	}
}

/* Parse reads the dna file and returns a DnaFileRecord object */
func (reader *DnaFileReader) Parse() (*DnaFileRecord, error) {
	file, err := os.Open(reader.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	record := newDnaFileRecord(reader.filePath)

	err = reader.parseHeader(file, record)
	if err != nil {
		return nil, err
	}

	for {
		var block byte
		err = binary.Read(file, binary.BigEndian, &block)
		if err != nil {
			break
		}

		var blockSize uint32
		err = binary.Read(file, binary.BigEndian, &blockSize)
		if err != nil {
			break
		}

		switch block {
		case 0:
			reader.parseSeqProperties(blockSize, file, record)
		case 6:
			data, err := reader.getBlockData(blockSize, file)
			if err != nil {
				return nil, err
			}
			err = reader.parseNotes(data, record)
			if err != nil {
				return nil, err
			}
		case 10:
			data, err := reader.getBlockData(blockSize, file)
			if err != nil {
				return nil, err
			}
			err = reader.parseFeatures(data, record)
			if err != nil {
				return nil, err
			}

		default:
			fmt.Println("Skipping block", block, "of size", blockSize)

			// Skip this block
			_, err = file.Seek(int64(blockSize), 1)
			if err != nil {
				return nil, err
			}
		}
	}

	return record, nil
}

func (reader *DnaFileReader) parseHeader(file *os.File, record *DnaFileRecord) error {
	var firstByte byte
	err := binary.Read(file, binary.BigEndian, &firstByte)
	if err != nil || firstByte != '\t' {
		return fmt.Errorf("first byte error. file is in an incorrect format")
	}

	var documentLength uint32
	err = binary.Read(file, binary.BigEndian, &documentLength)
	if err != nil || documentLength != 14 {
		return fmt.Errorf("document length is not 14")
	}

	title := make([]byte, 8)
	err = binary.Read(file, binary.BigEndian, &title)
	if err != nil || !bytes.Equal(title, []byte("SnapGene")) {
		return fmt.Errorf("document title is incorrect")
	}

	var isDNA bool
	var dnaFlag uint16
	err = binary.Read(file, binary.BigEndian, &dnaFlag)
	if err != nil {
		return err
	}
	isDNA = dnaFlag == 1

	var exportVersion, importVersion uint16
	err = binary.Read(file, binary.BigEndian, &exportVersion)
	if err != nil {
		return err
	}
	err = binary.Read(file, binary.BigEndian, &importVersion)
	if err != nil {
		return err
	}

	record.Meta = Meta{
		IsDna:         isDNA,
		ExportVersion: uint(exportVersion),
		ImportVersion: uint(importVersion),
	}

	return nil
}

func (sg *DnaFileReader) getBlockData(blockSize uint32, file *os.File) ([]byte, error) {
	content := make([]byte, blockSize)
	_, err := file.Read(content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

// Parser for the Notes block

// UnmarshalXML custom unmarshals the Notes structure to fill the map.
func unmarshalNotes(d *xml.Decoder) (map[string]string, error) {
	notes := make(map[string]string)
	for {
		// Read the next token from the decoder
		token, err := d.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return notes, err
		}

		fmt.Println(token)

		// If it's an end element matching the start element, break the loop
		switch elem := token.(type) {
		case xml.EndElement:
			if elem.Name.Local == "Notes" {
				return notes, nil
			}
		case xml.StartElement:
			if elem.Name.Local == "Notes" {
				continue
			}

			var value string
			err := d.DecodeElement(&value, &elem)
			if err != nil {
				return nil, err
			}
			notes[elem.Name.Local] = value
		}
	}
	return notes, nil
}

func (reader *DnaFileReader) parseNotes(block []byte, record *DnaFileRecord) error {
	notes, err := unmarshalNotes(
		xml.NewDecoder(bytes.NewReader(block)))
	if err != nil {
		return err
	}

	fmt.Println(notes)
	record.Notes = notes
	return nil
}

func (reader *DnaFileReader) parseFeatures(block []byte, record *DnaFileRecord) error {
	var features xmlFeatures
	err := xml.Unmarshal(block, &features)
	if err != nil {
		return err
	}

	resultFeatures := make([]Feature, 0)

	for _, feature := range features.FeatureList {

		segments := make([]FeatureSegment, 0)
		for _, segment := range feature.Segments {
			//fmt.Printf("  Segment Range: %s, Color: %s---%s\n", segment.Range, segment.Color, segment.Translated)

			start, end, err := parseRange(segment.Range)
			if err != nil {
				return err
			}

			var isTranslated = false
			if segment.Translated == "1" {
				isTranslated = true
			}

			segments = append(segments, FeatureSegment{
				Name:         segment.Name,
				Color:        segment.Color,
				Start:        start,
				End:          end,
				IsTranslated: isTranslated,
			})
		}

		qualifiers := make(map[string]string, 0)
		for _, q := range feature.Qualifiers {
			qualifiers[q.Name] = q.Value.Text
		}

		var label string
		if qualLabel, ok := qualifiers["label"]; ok {
			label = qualLabel
		} else {
			label = feature.Name
		}

		var start, end uint
		if len(segments) > 0 {
			// TODO: What do we do if there are multiple segments?
			// For now, just take the first segment
			start = segments[0].Start
			end = segments[0].End
		}

		newFeature := Feature{
			Name:  feature.Name,
			Type:  feature.Type,
			Label: label,

			Start: start,
			End:   end,

			Qualifiers: qualifiers,
			Segments:   segments,
		}

		resultFeatures = append(resultFeatures, newFeature)
	}

	record.Features = resultFeatures

	return nil
}

func parseRange(rangeStr string) (uint, uint, error) {
	var start, end uint
	_, err := fmt.Sscanf(rangeStr, "%d-%d", &start, &end)
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

/*

TODO: PARSE PRIMERS

case 5:

	/*root, err := sg.GetXML(blockSize, file)
	if err != nil {
		return err
	}*
	fmt.Println("NOT IMPLEMENTED")
	errors.New("Not implemented")
	//sg.ParsePrimers(root)

*


	/*
		func (sg *DnaFileRecord) ParsePrimers(root *xml.Element) {
			for _, primer := range root.Children {
				if primer.Tag == "Primer" {
					var pri Primer
					// Assuming Primer has a method FromElement to parse from xml.Element
					pri.FromElement(primer)
					sg.Primers = append(sg.Primers, pri)
				}
			}
		}
*/

func (reader *DnaFileReader) parseSeqProperties(blockSize uint32, file *os.File, record *DnaFileRecord) {
	var p byte
	err := binary.Read(file, binary.BigEndian, &p)
	if err != nil {
		return
	}

	properties := SequenceProperties{}

	properties.Topology = "linear"
	if p&0x01 > 0 {
		properties.Topology = "circular"
	}

	properties.Stranded = "single"
	if p&0x02 > 0 {
		properties.Stranded = "double"
	}

	properties.AMethylated = p&0x04 > 0
	properties.CMethylated = p&0x08 > 0
	properties.KiMethylated = p&0x10 > 0

	properties.Length = blockSize - 1

	seq := make([]byte, properties.Length)
	_, err = file.Read(seq)
	if err != nil {
		return
	}
	record.Sequence = string(seq)
	record.SequenceProperties = properties
}
