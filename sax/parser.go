package sax

import (
	"encoding/xml"
	"io"
)

// EventHandler is the interface definig the methods that handle relevant SAX
// parsing of an XML document.
type EventHandler interface {
	// OnStartDocument is invoked before starting the parsing.
	OnStartDocument() error

	// OnProcessingInstruction is called whenever a processing instruction (e.g.
	// <?xml ... ?>) is encountered.
	OnProcessingInstruction(element xml.ProcInst) error

	// OnStartElement is called whenever an open tag is encountered.
	OnStartElement(element xml.StartElement) error

	// OnEndElement is called whener an end tag is encountered; empty tags (e.g.
	// <tag/> generate a pair of events: OnStartElement and OnEndElement).
	OnEndElement(element xml.EndElement) error

	// OnCharacterData is called when characters are encountered out of tags; it
	// is invoked also on new lines,
	OnCharacterData(element xml.CharData) error

	// OnComment is invoked whenever there is a comment delimited by <!-- and -->;
	// the delimiters themselves are omitted.
	OnComment(element xml.Comment) error

	// OnEndDocument is called when the XML parsing is wrapping up.
	OnEndDocument() error
}

// ErrorHandler is the interface defining the methods that handle SAX parsing
// and other errors that might be raised while scanning an XML document.
type ErrorHandler interface {
	// OnError is invoked whenever there is an error (of any kind) while parsing
	// the XML document; if the method returns nil, the error is effectively
	// suppressed and the processing can go on; if a non-nil error is returned
	// (whether the original one or a new one) the Parser aborts the processing.
	OnError(err error) error
}

// Parser is an implementation of a SAX parser.
type Parser struct {
	EventHandler EventHandler
	ErrorHandler ErrorHandler
}

// Parse parses an XML document and invokes the SAX handlers' methods.
func (p *Parser) Parse(reader io.Reader) error {
	var err error
	d := xml.NewDecoder(reader)
	p.EventHandler.OnStartDocument()
loop:
	for {
		token, err := d.Token()
		switch {
		case err == io.EOF && token == nil:
			// done reading the document
			err = p.EventHandler.OnEndDocument()
			break loop
		case err != nil:
			// error reading the input XML, invoking error handler
			if p.ErrorHandler != nil {
				err = p.ErrorHandler.OnError(err)
				if err != nil {
					// error handler confirmed the error, aborting
					break loop
				} else {
					// error handler suppressed the error, we can continue
				}
			} else {
				// no error handler installed, bailing out with the parser error
				break loop
			}
		default:
			switch token := token.(type) {
			case xml.StartElement:
				p.EventHandler.OnStartElement(token.Copy())
			case xml.CharData:
				p.EventHandler.OnCharacterData(token.Copy())
			case xml.EndElement:
				p.EventHandler.OnEndElement(token)
			case xml.Comment:
				p.EventHandler.OnComment(token.Copy())
			case xml.ProcInst:
				p.EventHandler.OnProcessingInstruction(token.Copy())
			default:
				// unsupported token type???
			}
		}
	}
	return err
}

// DefaultHandler is the default, do-nothing implementation of the EventHandler
// and ErrorHandler interfaces.
type DefaultHandler struct{}

// OnStartDocument is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *DefaultHandler) OnStartDocument() error {
	return nil
}

// OnProcessingInstruction is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *DefaultHandler) OnProcessingInstruction(element xml.ProcInst) error {
	return nil
}

// OnStartElement is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *DefaultHandler) OnStartElement(element xml.StartElement) error {
	return nil
}

// OnEndElement is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *DefaultHandler) OnEndElement(element xml.EndElement) error {
	return nil
}

// OnCharacterData is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *DefaultHandler) OnCharacterData(element xml.CharData) error {
	return nil
}

// OnComment is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *DefaultHandler) OnComment(element xml.Comment) error {
	return nil
}

// OnEndDocument is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *DefaultHandler) OnEndDocument() error {
	return nil
}

// OnError is the default implementation of the corresponding ErrorHandler
// interface; it simply forwards any error to the Parser.
func (h *DefaultHandler) OnError(err error) error {
	return err
}
