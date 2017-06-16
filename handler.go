package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/dihedron/jted/stack"
)

// Handler is an implementation of the sax.EventHandler and sax.ErrorHandler
// interfaces.
type Handler struct {
	stack      *stack.Stack
	data       string
	buffer     bytes.Buffer
	all        bool              // if even empty tags should be parameterised
	template   bytes.Buffer      // where the template goes
	parameters map[string]string // where the parameters go
	tpl        io.Writer         // the (possibly nil) template file
	hcl        io.Writer         // the HCL file
}

// Close closes the underlying file descriptors.
func (h *Handler) Close() {
	if h.hcl != nil {
		defer h.hcl.(*bufio.Writer).Flush()
	}
	if h.tpl != nil {
		defer h.tpl.(*bufio.Writer).Flush()
	}
}

// OnStartDocument clears all data structures and gets ready for parsing a new
// XML document; this includes opening a file for the HCL example and, optionally,
// a file for the template if it is not embedded/inlined in the HCL file; the naming
// convention is the following: both files have the same base name as the config.xml
// with the .hcl and .tpl extensions.
func (h *Handler) OnStartDocument() error {
	h.stack.Clear()
	h.data = ""
	h.parameters = map[string]string{}
	h.buffer.Reset()
	h.template.Reset()
	return nil
}

// OnProcessingInstruction simply prints out the processing instructions as is.
func (h *Handler) OnProcessingInstruction(element xml.ProcInst) error {
	h.template.WriteString(fmt.Sprintf("<?%s %s?>\n", element.Target, string(element.Inst)))
	fmt.Printf("<?%s %s?>\n", element.Target, string(element.Inst))
	return nil
}

// OnStartElement pushes the element onto the stak; if the element is not
// the first on the stack, it marks its parent element, currently at the
// top of the stack, as a "container" so it can be treated accordingly: it
// will never be collapsed to a <tag/> because it is not empty.
func (h *Handler) OnStartElement(element xml.StartElement) error {
	// if there is already an open tag on the stack, then the current tag
	// is contained therein and we can mark the top of the stack as a
	// "container" node; this allows us to treat the parent node as a <tag>
	// </tag> pair and NEVER collapse it to <tag/>, which we only do for
	// empty leaf tags.
	if h.stack.Top() != nil && !h.stack.Top().(*Node).container {
		h.stack.Top().(*Node).container = true

		var buffer bytes.Buffer
		// if the parent node has attributes, format them, then print it out
		if len(h.stack.Top().(*Node).xml.(xml.StartElement).Attr) > 0 {
			for _, attr := range h.stack.Top().(*Node).xml.(xml.StartElement).Attr {
				buffer.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
			}
		}
		h.template.WriteString(fmt.Sprintf("%s<%s%s>\n", tab(h.stack.Len()-1), h.stack.Top().(*Node).xml.(xml.StartElement).Name.Local, buffer.String()))
		fmt.Printf("%s<%s%s>\n", tab(h.stack.Len()-1), bold(h.stack.Top().(*Node).xml.(xml.StartElement).Name.Local), buffer.String())
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
		if pattern.MatchString(h.data) {
			// if the value has already been parameterised "by hand", dump it as is
			h.template.WriteString(fmt.Sprintf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), h.data, element.Name.Local))
			h.parameters[h.data] = "<value>"
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), green(h.data), bold(element.Name.Local))
		} else {
			// otherwise calculate the name of the parameter
			parameter := templatise(top.Name.Local)
			h.template.WriteString(fmt.Sprintf("%s<%s%s>{{- %s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), parameter, element.Name.Local))
			h.parameters[parameter] = h.data
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), green("{{- "+parameter+" -}}"), bold(element.Name.Local))
		}
		h.data = ""
	} else if h.stack.Top() != nil && h.stack.Top().(*Node).container {
		h.template.WriteString(fmt.Sprintf("%s</%s>\n", tab(h.stack.Len()-1), element.Name.Local))
		fmt.Printf("%s</%s>\n", tab(h.stack.Len()-1), bold(element.Name.Local))
	} else {
		if h.all {
			parameter := templatise(top.Name.Local)
			h.template.WriteString(fmt.Sprintf("%s<%s%s>{{- %s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), parameter, element.Name.Local))
			h.parameters[parameter] = "<value>"
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), green("{{- "+parameter+" -}}"), bold(element.Name.Local))
		} else {
			h.template.WriteString(fmt.Sprintf("%s<%s%s/>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String()))
			fmt.Printf("%s<%s%s/>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String())
		}
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
	if h.tpl != nil {
		fmt.Fprint(h.tpl, h.template.String())
	}

	for k, v := range h.parameters {
		fmt.Printf("%q => %q\n", k, v)
	}
	return nil
}

// OnError is the default implementation of the corresponding ErrorHandler
// interface; it simply forwards any error to the Parser.
func (h *Handler) OnError(err error) error {
	return err
}
