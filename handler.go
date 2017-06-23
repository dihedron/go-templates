package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/dihedron/jted/stack"
)

// Handler is an implementation of the sax.EventHandler and sax.ErrorHandler
// interfaces.
type Handler struct {
	IncludeEmptyValues bool              // if even empty tags should be parameterised
	EmbedConfigXML     bool              // if the confg.xml template should be inlined
	ConfigXML          bytes.Buffer      // the buffer where the config.xml template goes
	HCL                bytes.Buffer      // the buffer where the HCL goes
	stack              *stack.Stack      // the SAX internal stack
	currentValue       string            // the value of the current parameter
	parameters         map[string]string // where the parameters go
}

// OnStartDocument clears all data structures and gets ready for parsing a new
// XML document; this includes opening a file for the HCL example and, optionally,
// a file for the template if it is not embedded/inlined in the HCL file; the naming
// convention is the following: both files have the same base name as the config.xml
// with the .hcl and .tpl extensions.
func (h *Handler) OnStartDocument() error {
	h.stack.Clear()
	h.currentValue = ""

	h.parameters = map[string]string{}
	h.HCL.Reset()
	h.HCL.WriteString(`
\*
 * Jenkins job definition
 */
resource "jenkins_job" "<job name here>" {	
    name                               = "<job name here>"
    display_name                       = "<[optional] job display name here>"
    description                        = "<job description here>"
    disabled                           = false	
`)
	h.ConfigXML.Reset()
	return nil
}

// OnProcessingInstruction simply prints out the processing instructions as is.
func (h *Handler) OnProcessingInstruction(element xml.ProcInst) error {
	h.ConfigXML.WriteString(fmt.Sprintf("<?%s %s?>\n", element.Target, string(element.Inst)))
	return nil
}

// OnStartElement pushes the element onto the stack; if the element is not
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
		h.ConfigXML.WriteString(fmt.Sprintf("%s<%s%s>\n", tab(h.stack.Len()-1), h.stack.Top().(*Node).xml.(xml.StartElement).Name.Local, buffer.String()))
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
	if len(h.currentValue) > 0 {
		if pattern.MatchString(h.currentValue) {
			// if the value has already been parameterised "by hand", dump it as is
			h.ConfigXML.WriteString(fmt.Sprintf("%s<%s%s>%s</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), h.currentValue, element.Name.Local))
			h.parameters[h.currentValue] = "<no value provided>"
		} else {
			// otherwise calculate the name of the parameter
			parameter := templatise(top.Name.Local)
			if isSpecialParameter(parameter) {
				// if it is one of the "top level", special paramweters we do not prefix
				// it with ".parameters" and we do not capitalise it (use original form)
				h.ConfigXML.WriteString(fmt.Sprintf("%s<%s%s>{{- .%s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), top.Name.Local, element.Name.Local))
			} else {
				h.ConfigXML.WriteString(fmt.Sprintf("%s<%s%s>{{- .parameters.%s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), parameter, element.Name.Local))
			}
			h.parameters[parameter] = h.currentValue
		}
		h.currentValue = ""
	} else if h.stack.Top() != nil && h.stack.Top().(*Node).container {
		h.ConfigXML.WriteString(fmt.Sprintf("%s</%s>\n", tab(h.stack.Len()-1), element.Name.Local))
	} else {
		if h.IncludeEmptyValues {
			parameter := templatise(top.Name.Local)
			if isSpecialParameter(parameter) {
				// if it is one of the "top level", special paramweters we do not prefix
				// it with ".parameters" and we do not capitalise it (use original form)
				h.ConfigXML.WriteString(fmt.Sprintf("%s<%s%s>{{- .%s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), top.Name.Local, element.Name.Local))
			} else {
				h.ConfigXML.WriteString(fmt.Sprintf("%s<%s%s>{{- .parameters.%s -}}</%s>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String(), parameter, element.Name.Local))
			}
			h.parameters[parameter] = "<no value provided>"
		} else {
			h.ConfigXML.WriteString(fmt.Sprintf("%s<%s%s/>\n", tab(h.stack.Len()-1), top.Name.Local, buffer.String()))
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
		h.currentValue = data
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
	if len(h.parameters) > 0 {
		h.HCL.WriteString(fmt.Sprintf("\t%-40s= {\n", "parameters"))
		for k, v := range h.parameters {
			if _, err := strconv.ParseInt(v, 10, 64); err == nil {
				h.HCL.WriteString(fmt.Sprintf("\t\t%-36s= %s,\n", k, v))
			} else if b, err := strconv.ParseBool(v); err == nil {
				h.HCL.WriteString(fmt.Sprintf("\t\t%-36s= %t,\n", k, b))
			} else {
				h.HCL.WriteString(fmt.Sprintf("\t\t%-36s= \"%s\",\n", k, v))
			}
		}
		h.HCL.WriteString("\t}\n")
	}
	if h.EmbedConfigXML {
		// config.xml template must be inlined
		h.HCL.WriteString(fmt.Sprintf("\t%-40s=<<EOF\n", "template"))
		h.HCL.WriteString(h.ConfigXML.String())
		h.HCL.WriteString("EOF\n")
	} else {
		h.HCL.WriteString(fmt.Sprintf("\t%-40s= \"file://%%s\"\n", "template"))
	}
	h.HCL.WriteString("}")
	return nil
}

// OnError is the default implementation of the corresponding ErrorHandler
// interface; it simply forwards any error to the Parser.
func (h *Handler) OnError(err error) error {
	return err
}

func isSpecialParameter(name string) bool {
	// Name is treated differently because it is (or should) never be in the config XML
	// and is usually sent to the server in the POST request; it appears in some
	// configuration tags though, so it mnust be treated as an ordinary parameter
	for _, special := range []string{ /*"Name",*/ "DisplayName", "Disabled", "Template", "Description"} {
		if name == special {
			return true
		}
	}
	return false
}
