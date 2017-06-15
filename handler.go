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

// Handler is the implementation of the sax.EventHandler and sax.ErrorHandler
// interfaces.
type Handler struct {
	stack  *stack.Stack
	data   string
	buffer bytes.Buffer
	all    bool
	tpl    io.Writer
	hcl    io.Writer
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
// a file for the template it is not embedded/inlined in the HCL file; the naming
// convention is the following: both files have the same base name as the config.xml
// with the .hcl and .tpl extensions.
func (h *Handler) OnStartDocument() error {
	h.stack.Clear()
	h.data = ""
	h.buffer.Reset()
	return nil
}

// OnProcessingInstruction is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnProcessingInstruction(element xml.ProcInst) error {
	fmt.Printf("<?%s %s?>\n", element.Target, string(element.Inst))
	fmt.Fprintf(h.tpl, "<?%s %s?>\n", element.Target, string(element.Inst))
	return nil
}

// OnStartElement is the default, do-nothing implementation of the corresponding
// EventHandler interface.
func (h *Handler) OnStartElement(element xml.StartElement) error {
	// if there is already an open tag on the stack, then the current tag
	// is contained therein and we can mark the top of the stack as a
	// "container" node; this allows us to treat the parent node as a <tag>
	// </tag> pair and NEVER collapse it to <tag/>, which we only do for
	// empty leaf tags.
	if h.stack.Top() != nil && !h.stack.Top().(*Node).container {
		h.stack.Top().(*Node).container = true
		var buffer bytes.Buffer
		// if the current node has attributes, format them
		if len(h.stack.Top().(*Node).xml.(xml.StartElement).Attr) > 0 {
			for _, attr := range h.stack.Top().(*Node).xml.(xml.StartElement).Attr {
				buffer.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
			}
		}
		fmt.Fprintf(h.tpl, "%s<%s%s>\n", tab(h.stack.Len()-1), h.stack.Top().(*Node).xml.(xml.StartElement).Name.Local, buffer.String())
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
			fmt.Fprintf(h.tpl, "%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), h.data, element.Name.Local)
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), green(h.data), bold(element.Name.Local))
			// TODO: output to HCL too
		} else {
			// otherwise calculate the name of the parameter
			parameter := templatise(top.Name.Local)
			// TODO: insert into map parameter -> h.data
			fmt.Fprintf(h.tpl, "%s<%s%s>{{- %s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), parameter, element.Name.Local)
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), green("{{- "+parameter+" -}}"), bold(element.Name.Local))
			// TODO: output to HCL too
		}
		h.data = ""
	} else if h.stack.Top() != nil && h.stack.Top().(*Node).container {
		fmt.Fprintf(h.tpl, "%s</%s>\n", tab(h.stack.Len()-1), element.Name.Local)
		fmt.Printf("%s</%s>\n", tab(h.stack.Len()-1), bold(element.Name.Local))
	} else {
		if h.all {
			parameter := templatise(top.Name.Local)
			fmt.Fprintf(h.tpl, "%s<%s%s>{{- %s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), parameter, element.Name.Local)
			fmt.Printf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), bold(top.Name.Local), buffer.String(), green("{{- "+parameter+" -}}"), bold(element.Name.Local))

		} else {
			//log.Printf("%s<%s%s/>\n", tab(h.stack.Len()-1), element.Name.Local, buffer.String())
			fmt.Fprintf(h.tpl, "%s<%s%s/>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String())
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
	return nil
}

// OnError is the default implementation of the corresponding ErrorHandler
// interface; it simply forwards any error to the Parser.
func (h *Handler) OnError(err error) error {
	return err
}
