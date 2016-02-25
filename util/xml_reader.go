package util

import (
	"encoding/xml"
	"os"
)

// idea and sample code for the XML stream reader from here:
//   http://blog.davidsingleton.org/parsing-huge-xml-files-with-go/
// XmlReader was implemented for GoGrinder-samples/xmlcowboys

func XmlReader(filename string, element string) <-chan string {
	read := make(chan string)

	file, err := os.Open(filename) // For read access.
	if err != nil {
		panic(err)
	}

	type Inner struct {
		Value string `xml:",innerxml"`
	}

	decoder := xml.NewDecoder(file)

	go func() {
		for {
			// Read tokens from the XML document in a stream.
			t, _ := decoder.Token()
			if t == nil {
				break
			}
			// Inspect the type of the token just read.
			switch se := t.(type) {
			case xml.StartElement:
				// If we just read a StartElement token
				// ...and it matches the <element> we are looking for
				if se.Name.Local == element {
					var e Inner
					// decode a whole chunk of following XML into the
					// variable r which is Inner struct
					decoder.DecodeElement(&e, &se)
					read <- e.Value
				}
			}
		}
		close(read)
	}()
	return read
}
