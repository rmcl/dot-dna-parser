package dnafile

import "encoding/xml"

type DnaFileRecord struct {
	FilePath      string
	Primers       []Primer
	Features      map[string][]Feature
	NotesContent  []Note
	SeqProperties map[string]interface{}
	Meta          map[string]interface{}
	Translation   string
}

type Note struct {
	Tag     string
	Content string
}

type Features struct {
	XMLName     xml.Name  `xml:"Features"`
	NextValidID string    `xml:"nextValidID,attr"`
	FeatureList []Feature `xml:"Feature"`
}

type Feature struct {
	RecentID                        string    `xml:"recentID,attr"`
	Name                            string    `xml:"name,attr"`
	Directionality                  string    `xml:"directionality,attr"`
	TranslationMW                   string    `xml:"translationMW,attr,omitempty"`
	Type                            string    `xml:"type,attr"`
	SwappedSegmentNumbering         string    `xml:"swappedSegmentNumbering,attr"`
	AllowSegmentOverlaps            string    `xml:"allowSegmentOverlaps,attr"`
	CleavageArrows                  string    `xml:"cleavageArrows,attr,omitempty"`
	ReadingFrame                    string    `xml:"readingFrame,attr,omitempty"`
	ConsecutiveTranslationNumbering string    `xml:"consecutiveTranslationNumbering,attr"`
	HitsStopCodon                   string    `xml:"hitsStopCodon,attr,omitempty"`
	DetectionMode                   string    `xml:"detectionMode,attr,omitempty"`
	Segments                        []Segment `xml:"Segment"`
	Qs                              []Q       `xml:"Q"`
}

type Segment struct {
	Name       string `xml:"name,attr,omitempty"`
	Range      string `xml:"range,attr"`
	Color      string `xml:"color,attr"`
	Type       string `xml:"type,attr"`
	Translated string `xml:"translated,attr,omitempty"`
}

type Q struct {
	Name string `xml:"name,attr"`
	V    V      `xml:"V"`
}

type V struct {
	Int  int    `xml:"int,attr,omitempty"`
	Text string `xml:"text,attr,omitempty"`
}
