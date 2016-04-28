package util

import (
    //"compress/gzip"
	"encoding/xml"
	"os"
    "io"
    "compress/gzip"
)

// idea and sample code for the XML stream reader from here:
//   http://blog.davidsingleton.org/parsing-huge-xml-files-with-go/
// XmlReader was implemented for GoGrinder-samples/xmlcowboys

// Read lines from an XML file
func xmlReader(fi io.Reader, element string) <-chan string {
    read := make(chan string)

    type Inner struct {
        Value string `xml:",innerxml"`
    }

    decoder := xml.NewDecoder(fi)

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


// Read lines from an XML file
func XmlReader(filename string, element string) <-chan string {
    fi, err := os.Open(filename) // For read access.
    if err != nil {
        panic(err)
    }
    defer fi.Close()

    return xmlReader(fi, element)
}


// Read lines from an GZipped XML file
func GzXmlReader(filename string, element string) <-chan string {
    fi, err := os.Open(filename) // For read access.
    if err != nil {
        panic(err)
    }
    defer fi.Close()
    gr, err := gzip.NewReader(fi)
    if err != nil {
        panic(err)
    }
    defer gr.Close()

    return xmlReader(gr, element)
}
