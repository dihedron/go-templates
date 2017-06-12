package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/dihedron/jted/stack"
)

// PrintHandler is the implementation of the sax.EventHandler and sax.ErrorHandler
// interfaces.
type PrintHandler struct {
	stack  *stack.Stack
	data   string
	buffer bytes.Buffer
}

// OnStartDocument clears all data structures and gets ready for parsing a new
// XML document.
func (h *PrintHandler) OnStartDocument() error {
	h.stack.Clear()
	h.data = ""
	h.buffer.Reset()
	return nil
}

// OnProcessingInstruction is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *PrintHandler) OnProcessingInstruction(element xml.ProcInst) error {
	fmt.Printf("<?%s %s?>\n", element.Target, string(element.Inst))
	return nil
}

// OnStartElement is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *PrintHandler) OnStartElement(element xml.StartElement) error {
	if h.stack.Top() != nil && !h.stack.Top().(*Node).container {
		h.stack.Top().(*Node).container = true
		var buffer bytes.Buffer
		if len(h.stack.Top().(*Node).xml.(xml.StartElement).Attr) > 0 {
			for _, attr := range h.stack.Top().(*Node).xml.(xml.StartElement).Attr {
				if pattern.MatchString(attr.Value) {
					buffer.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, green(attr.Value)))
				} else {
					buffer.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
				}

			}
		}
		fmt.Printf("%s<%s%s>\n", tab(h.stack.Len()-1), bold(h.stack.Top().(*Node).xml.(xml.StartElement).Name.Local), buffer.String())
	}
	h.stack.Push(&Node{xml: element})
	return nil
}

// OnEndElement is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *PrintHandler) OnEndElement(element xml.EndElement) error {
	top := h.stack.Top().(*Node).xml.(xml.StartElement)
	var buffer bytes.Buffer
	if len(h.stack.Top().(*Node).xml.(xml.StartElement).Attr) > 0 {
		for _, attr := range h.stack.Top().(*Node).xml.(xml.StartElement).Attr {
			buffer.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
		}
	}
	if len(h.data) > 0 {
		if pattern.MatchString(h.data) {
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), green(h.data), bold(element.Name.Local))
		} else {
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), red(h.data), bold(element.Name.Local))
		}
		h.data = ""
	} else if h.stack.Top() != nil && h.stack.Top().(*Node).container {
		fmt.Printf("%s</%s>\n", tab(h.stack.Len()-1), bold(element.Name.Local))
	} else {
		//log.Printf("%s<%s%s/>\n", tab(h.stack.Len()-1), element.Name.Local, buffer.String())
		fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), red("???"), bold(element.Name.Local))
	}
	h.stack.Pop()
	return nil
}

// OnCharacterData is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *PrintHandler) OnCharacterData(element xml.CharData) error {
	data := strings.TrimSpace(string(element))
	if len(data) > 0 {
		h.data = data
	}
	return nil
}

// OnComment is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *PrintHandler) OnComment(element xml.Comment) error {
	return nil
}

// OnEndDocument is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *PrintHandler) OnEndDocument() error {
	return nil
}

// OnError is the default implementation of the corresponding ErrorHandler
// interface; it simply forwards any error to the Parser.
func (h *PrintHandler) OnError(err error) error {
	return err
}
