package dnafile

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"os"
)

type Primer struct {
	// Define the fields based on your Primer class structure
}

func NewDnaFileReader(filePath string) *DnaFileRecord {
	return &DnaFileRecord{
		FilePath:      filePath,
		Primers:       []Primer{},
		Features:      make(map[string][]Feature),
		NotesContent:  []Note{},
		SeqProperties: make(map[string]interface{}),
		Meta:          make(map[string]interface{}),
	}
}

func (sg *DnaFileRecord) Parse() error {
	file, err := os.Open(sg.FilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	err = sg.parseHeader(file)
	if err != nil {
		return err
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

		fmt.Println("Block:", block, "Size:", blockSize)

		switch block {
		case 0:
			sg.ParseSeqProperties(blockSize, file)
		case 6:
			data, err := sg.getBlockData(blockSize, file)
			if err != nil {
				return err
			}
			err = sg.ParseNotes(data)
			if err != nil {
				return err
			}
		case 10:
			data, err := sg.getBlockData(blockSize, file)
			if err != nil {
				return err
			}
			err = sg.ParseFeatures(data)
			if err != nil {
				return err
			}

		default:
			fmt.Println("Skipping block", block, "of size", blockSize)

			// Skip this block
			_, err = file.Seek(int64(blockSize), 1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (sg *DnaFileRecord) parseHeader(file *os.File) error {
	var firstByte byte
	err := binary.Read(file, binary.BigEndian, &firstByte)
	if err != nil || firstByte != '\t' {
		return fmt.Errorf("file is in an incorrect format")
	}

	var documentLength uint32
	err = binary.Read(file, binary.BigEndian, &documentLength)
	if err != nil || documentLength != 14 {
		return fmt.Errorf("file is in an incorrect format")
	}

	title := make([]byte, 8)
	err = binary.Read(file, binary.BigEndian, &title)
	if err != nil || !bytes.Equal(title, []byte("DnaFileRecord")) {
		return fmt.Errorf("file is in an incorrect format")
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

	sg.Meta = map[string]interface{}{
		"is_dna":         isDNA,
		"export_version": exportVersion,
		"import_version": importVersion,
	}
	return nil
}

func (sg *DnaFileRecord) getBlockData(blockSize uint32, file *os.File) ([]byte, error) {
	content := make([]byte, blockSize)
	_, err := file.Read(content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (sg *DnaFileRecord) ParseNotes(block []byte) error {

	var notes []Note
	err := xml.Unmarshal(block, &notes)
	if err != nil {
		return err
	}

	sg.NotesContent = notes
	return nil
}

func (sg *DnaFileRecord) ParseFeatures(block []byte) error {
	var features Features
	err := xml.Unmarshal(block, &features)
	if err != nil {
		fmt.Println("Error unmarshalling XML:", err)
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

func (sg *DnaFileRecord) ParseSeqProperties(blockSize uint32, file *os.File) {
	var p byte
	err := binary.Read(file, binary.BigEndian, &p)
	if err != nil {
		return
	}

	sg.SeqProperties["topology"] = "linear"
	if p&0x01 > 0 {
		sg.SeqProperties["topology"] = "circular"
	}

	sg.SeqProperties["stranded"] = "single"
	if p&0x02 > 0 {
		sg.SeqProperties["stranded"] = "double"
	}

	sg.SeqProperties["a_methylated"] = p&0x04 > 0
	sg.SeqProperties["c_methylated"] = p&0x08 > 0
	sg.SeqProperties["ki_methylated"] = p&0x10 > 0

	seqLength := blockSize - 1
	sg.SeqProperties["length"] = seqLength

	seq := make([]byte, seqLength)
	_, err = file.Read(seq)
	if err != nil {
		return
	}
	sg.SeqProperties["seq"] = string(seq)
}
