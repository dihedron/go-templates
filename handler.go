package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strings"

	"github.com/dihedron/jted/stack"
)

// Handler is the implementation of the sax.EventHandler and sax.ErrorHandler
// interfaces.
type Handler struct {
	stack  *stack.Stack
	data   string
	buffer bytes.Buffer
}

// Node describes a node in the XML tree.
type Node struct {
	xml       interface{}
	container bool
}

// OnStartDocument clears all data structures and gets ready for parsing a new
// XML document.
func (h *Handler) OnStartDocument() error {
	h.stack.Clear()
	h.data = ""
	h.buffer.Reset()
	return nil
}

// OnProcessingInstruction is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnProcessingInstruction(element xml.ProcInst) error {
	log.Printf("<?%s %s?>\n", element.Target, string(element.Inst))
	return nil
}

// OnStartElement is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnStartElement(element xml.StartElement) error {
	if h.stack.Top() != nil && !h.stack.Top().(*Node).container {
		h.stack.Top().(*Node).container = true
		var buffer bytes.Buffer
		if len(h.stack.Top().(*Node).xml.(xml.StartElement).Attr) > 0 {
			for _, attr := range h.stack.Top().(*Node).xml.(xml.StartElement).Attr {
				buffer.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
			}
		}
		log.Printf("%s<%s%s>\n", tab(h.stack.Len()-1), h.stack.Top().(*Node).xml.(xml.StartElement).Name.Local, buffer.String())
	}
	h.stack.Push(&Node{xml: element})
	return nil
}

// OnEndElement is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnEndElement(element xml.EndElement) error {
	top := h.stack.Top().(*Node).xml.(xml.StartElement)
	var buffer bytes.Buffer
	if len(h.stack.Top().(*Node).xml.(xml.StartElement).Attr) > 0 {
		for _, attr := range h.stack.Top().(*Node).xml.(xml.StartElement).Attr {
			buffer.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
		}
	}
	if len(h.data) > 0 {
		log.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), h.data, element.Name.Local)
		h.data = ""
	} else if h.stack.Top() != nil && h.stack.Top().(*Node).container {
		log.Printf("%s</%s>\n", tab(h.stack.Len()-1), element.Name.Local)
	} else {
		log.Printf("%s<%s%s/>\n", tab(h.stack.Len()-1), element.Name.Local, buffer.String())
	}
	h.stack.Pop()
	return nil
}

// OnCharacterData is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnCharacterData(element xml.CharData) error {
	data := strings.TrimSpace(string(element))
	if len(data) > 0 {
		h.data = data
	}
	return nil
}

// OnComment is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnComment(element xml.Comment) error {
	return nil
}

// OnEndDocument is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnEndDocument() error {
	return nil
}

// OnError is the default implementation of the corresponding ErrorHandler
// interface; it simply forwards any error to the Parser.
func (h *Handler) OnError(err error) error {
	return err
}

// tab creates a string with the given number of tabs; each tab has a size of
// two blank spaces in the current implementation.
func tab(count int) string {
	return fmt.Sprintf(fmt.Sprintf("%%-%ds", count*2), "")
}
