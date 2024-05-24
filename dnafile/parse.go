package dnafile

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
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
		Features:           make(map[string][]Feature),
		Notes:              make(map[string]string),
		SequenceProperties: SequenceProperties{},
		Meta:               make(map[string]interface{}),
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
			reader.ParseSeqProperties(blockSize, file, record)
		case 6:
			data, err := reader.getBlockData(blockSize, file)
			if err != nil {
				return nil, err
			}
			err = reader.ParseNotes(data, record)
			if err != nil {
				return nil, err
			}
		case 10:
			data, err := reader.getBlockData(blockSize, file)
			if err != nil {
				return nil, err
			}
			err = reader.ParseFeatures(data, record)
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

	record.Meta = map[string]interface{}{
		"is_dna":         isDNA,
		"export_version": exportVersion,
		"import_version": importVersion,
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
type notesXml struct {
	Fields map[string]string
}

// UnmarshalXML custom unmarshals the Notes structure to fill the map.
func (n *notesXml) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Fields = make(map[string]string)
	for {
		// Read the next token from the decoder
		token, err := d.Token()
		if err != nil {
			return err
		}

		// If it's an end element matching the start element, break the loop
		switch elem := token.(type) {
		case xml.EndElement:
			if elem.Name == start.Name {
				return nil
			}
		case xml.StartElement:
			var value string
			err := d.DecodeElement(&value, &elem)
			if err != nil {
				return err
			}
			n.Fields[elem.Name.Local] = value
		}
	}
}

func (reader *DnaFileReader) ParseNotes(block []byte, record *DnaFileRecord) error {

	var notes notesXml
	err := xml.NewDecoder(bytes.NewReader(block)).Decode(&notes)
	if err != nil {
		return err
	}

	record.Notes = notes.Fields
	return nil
}

func (reader *DnaFileReader) ParseFeatures(block []byte, record *DnaFileRecord) error {
	var features Features
	err := xml.Unmarshal(block, &features)
	if err != nil {
		return err
	}

	// Print the parsed data
	fmt.Printf("NextValidID: %s\n", features.NextValidID)
	for _, feature := range features.FeatureList {
		fmt.Printf("Feature Name: %s, Type: %s\n", feature.Name, feature.Type)
		for _, segment := range feature.Segments {
			fmt.Printf("  Segment Range: %s, Color: %s\n", segment.Range, segment.Color)
		}
		for _, q := range feature.Qs {
			fmt.Printf("  Q Name: %s, Value: %s\n", q.Name, q.V.Text)
		}
	}

	return nil
}

/*

			case 5:

				/*root, err := sg.GetXML(blockSize, file)
				if err != nil {
					return err
				}*
				fmt.Println("NOT IMPLEMENTED")
				errors.New("Not implemented")
				//sg.ParsePrimers(root)
			default:
				_, err = file.Seek(int64(blockSize), 1)
				if err != nil {
					return err
				}
			}
			*
		}

		return nil
	}

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

func (reader *DnaFileReader) ParseSeqProperties(blockSize uint32, file *os.File, record *DnaFileRecord) {
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
