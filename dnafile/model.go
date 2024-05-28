package dnafile

import "encoding/xml"

type DnaFileRecord struct {
	FilePath string
	//Primers       []Primer
	Features           []Feature
	Notes              map[string]string
	Sequence           string
	SequenceProperties SequenceProperties
	Meta               Meta
	Translation        string
}

type Meta struct {
	IsDna         bool
	ExportVersion uint
	ImportVersion uint
}

type Feature struct {
	Start  uint
	End    uint
	Strand string

	Type  string
	Name  string
	Label string

	Color     string
	TextColor string

	Segments []FeatureSegment

	Row   uint
	IsOrf bool

	Qualifiers map[string]string
}

type FeatureSegment struct {
	Name         string
	Color        string
	Start        uint
	End          uint
	IsTranslated bool
}

type SequenceProperties struct {
	Topology string
	Stranded string

	AMethylated  bool
	CMethylated  bool
	KiMethylated bool

	Length uint32
}

/* XML structs
 * These are used to unmarshal the XML data from the DNA file.
 */

type xmlFeatures struct {
	XMLName     xml.Name     `xml:"Features"`
	NextValidID string       `xml:"nextValidID,attr"`
	FeatureList []xmlFeature `xml:"Feature"`
}

type xmlFeature struct {
	RecentID                        string         `xml:"recentID,attr"`
	Name                            string         `xml:"name,attr"`
	Directionality                  string         `xml:"directionality,attr"`
	TranslationMW                   string         `xml:"translationMW,attr,omitempty"`
	Type                            string         `xml:"type,attr"`
	SwappedSegmentNumbering         string         `xml:"swappedSegmentNumbering,attr"`
	AllowSegmentOverlaps            string         `xml:"allowSegmentOverlaps,attr"`
	CleavageArrows                  string         `xml:"cleavageArrows,attr,omitempty"`
	ReadingFrame                    string         `xml:"readingFrame,attr,omitempty"`
	ConsecutiveTranslationNumbering string         `xml:"consecutiveTranslationNumbering,attr"`
	HitsStopCodon                   string         `xml:"hitsStopCodon,attr,omitempty"`
	DetectionMode                   string         `xml:"detectionMode,attr,omitempty"`
	Segments                        []xmlSegment   `xml:"Segment"`
	Qualifiers                      []xmlQualifier `xml:"Q"`
}

type xmlSegment struct {
	Name       string `xml:"name,attr,omitempty"`
	Range      string `xml:"range,attr"`
	Color      string `xml:"color,attr"`
	Type       string `xml:"type,attr"`
	Translated string `xml:"translated,attr,omitempty"`
}

type xmlQualifier struct {
	Name  string   `xml:"name,attr"`
	Value xmlValue `xml:"V"`
}

type xmlValue struct {
	Int  int    `xml:"int,attr,omitempty"`
	Text string `xml:"text,attr,omitempty"`
}

type xmlNotes struct {
	Name  string   `xml:"name,attr"`
	Value xmlValue `xml:"V"`
}
