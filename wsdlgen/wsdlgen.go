// Package wsdlgen generates Go source code from wsdl documents.
//
// The wsdlgen package generates Go source for calling the various
// methods defined in a WSDL (Web Service Definition Language) document.
// The generated Go source is self-contained, with no dependencies on
// non-standard packages.
//
// Code generation for the wsdlgen package can be configured by using
// the provided Option functions.
package wsdlgen // import "aqwari.net/xml/wsdlgen"

import (
	"encoding/xml"
	"errors"
	"fmt"
	"go/ast"
	"os"

	"aqwari.net/xml/internal/gen"
	"aqwari.net/xml/wsdl"
	"aqwari.net/xml/xsd"
	"aqwari.net/xml/xsdgen"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Types conforming to the Logger interface can receive information about
// the code generation process.
type Logger interface {
	Printf(format string, v ...interface{})
}

type printer struct {
	*Config
	code       *xsdgen.Code
	wsdl       *wsdl.Definition
	file       *ast.File
	decl       []ast.Decl
	schemas    []xsd.Schema
	wsdlschema xsd.Schema
}

// Provides aspects about an RPC call to the template for the function
// bodies.
type opArgs struct {
	// formatted with appropriate variable names
	input, output []string

	// URL to send request to
	Address string

	// POST or GET
	Method string

	SOAPAction string

	// true for "document" style, false for "rpc" encoding
	// https://schemas.xmlsoap.org/wsdl/soap/#tStyleChoice
	DocumentStyle bool

	// Name of the method to call
	MsgName xml.Name

	// if we're returning individual values, these slices
	// are in an order matching the input/output slices.
	InputName, OutputName xml.Name
	InputFields           []field
	OutputFields          []field

	// If not "", inputs come in a wrapper struct
	InputType string

	// If not "", we return values in a wrapper struct
	ReturnType   string
	ReturnFields []field
}

// struct members. Need to export the fields for our template
type field struct {
	Name, Type string
	XMLName    xml.Name

	// If this is a wrapper struct for >InputThreshold arguments,
	// PublicType holds the type that we want to expose to the
	// user. For example, if the web service expects an xsdDate
	// to be sent to it, PublicType will be time.Time and a conversion
	// will take place before sending the request to the server.
	PublicType string

	// This refers to the name of the value to assign to this field
	// in the argument list. Empty for return values.
	InputArg string
}

// GenAST creates a Go source file containing type and method declarations
// that can be used to access the service described in the provided set of wsdl
// files.
func (cfg *Config) GenAST(files ...string) (*ast.File, error) {
	if len(files) == 0 {
		return nil, errors.New("must provide at least one file name")
	}
	if cfg.pkgName == "" {
		cfg.pkgName = "ws"
	}
	if cfg.pkgHeader == "" {
		cfg.pkgHeader = fmt.Sprintf("Package %s", cfg.pkgName)
	}

	docs := make([][]byte, 0, len(files))
	for _, filename := range files {
		if data, err := os.ReadFile(filename); err != nil {
			return nil, err
		} else {
			cfg.debugf("read %s", filename)
			docs = append(docs, data)
		}
	}

	cfg.debugf("parsing WSDL file %s", files[0])
	def, err := wsdl.Parse(docs[0])
	if err != nil {
		return nil, err
	}

	cfg.verbosef("generating function definitions from WSDL")
	return cfg.genAST(def, docs)
}

func (cfg *Config) genAST(def *wsdl.Definition, docs [][]byte) (*ast.File, error) {
	cfg.verbosef("generating type declarations from xml schema")

	schemas, err := cfg.xsdgen.ParseSchemas(docs...)
	if err != nil {
		return nil, err
	}

	p := &printer{
		Config:  cfg,
		wsdl:    def,
		decl:    nil,
		schemas: schemas,
		wsdlschema: xsd.Schema{
			TargetNS: def.TargetNS,
			Types:    make(map[xml.Name]xsd.Type),
		},
	}
	//convert all RPC style arguments to xsd struct for document style use
	//this also applies all the overlays for handling non-trival basic types (e.g. xsd:date)
	p.genASTpre()
	{
		code, err := cfg.xsdgen.GenCodeWithSchema(&p.wsdlschema, docs...)
		if err != nil {
			return nil, err
		}
		file, err := code.GenAST()
		if err != nil {
			return nil, err
		}
		file.Name = ast.NewIdent(cfg.pkgName)
		file = gen.PackageDoc(file, cfg.pkgHeader, "\n", def.Doc)

		p.file = file
		p.code = code
	}

	p.genAST()

	p.file.Decls = append(p.file.Decls, p.decl...)
	//prepend import statement

	return p.file, nil
}

func (p *printer) genASTpre() error {
	for i := range p.wsdl.Ports {
		if err := p.portPre(&p.wsdl.Ports[i]); err != nil {
			return err
		}
	}
	return nil
}

// Preprocessor identifying structs that should be generated for rpc
// style operations
func (p *printer) portPre(port *wsdl.Port) error {
	for i := range port.Operations {
		if err := p.operationPre(&(port.Operations[i])); err != nil {
			return err
		}
	}
	return nil
}

func (p *printer) operationPre(op *wsdl.Operation) error {
	input, ok := p.wsdl.Message[op.Input]
	if !ok {
		return fmt.Errorf("unknown input message type %s", op.Input.Local)
	}
	output, ok := p.wsdl.Message[op.Output]
	if !ok {
		return fmt.Errorf("unknown output message type %s", op.Output.Local)
	}
	if !op.DocumentStyle {
		p.wsdl.Message[op.Input] = p.messageToComplexType(input)
		p.wsdl.Message[op.Output] = p.messageToComplexType(output)
		op.DocumentStyle = true
	}

	return nil
}

// convert wsdl message that is rpc style to complex type that inserts
// into the soap body in document style
func (p *printer) messageToComplexType(msg wsdl.Message) wsdl.Message {

	elements := []xsd.Element{}
	for _, pt := range msg.Parts {
		var foundType xsd.Type

		if bt, err := xsd.ParseBuiltin(pt.Type); err == nil {
			foundType = bt
		}
		if bt, err := xsd.ParseBuiltin(pt.Element); err == nil {
			foundType = bt
		}
		if bt, err := xsd.ParseBuiltin(xml.Name{Space: "http://www.w3.org/2001/XMLSchema", Local: pt.Element.Local}); err == nil {
			foundType = bt
		}
		for _, s := range p.schemas {
			if t := s.FindType(pt.Type); t != nil {
				foundType = t
			}
			if t := s.FindType(pt.Element); t != nil {
				foundType = t
			}
			if s.TargetNS == pt.Element.Space {
				for k, v := range s.Types {
					if k == pt.Element {
						foundType = v
					}
				}
			}
		}

		if foundType == nil {
			panic("wsdl parse error unimplemented wsdl type while parsing wsdl message " + msg.Name.Local + " " + pt.Element.Space + "#" + pt.Element.Local)
		}

		elements = append(elements, xsd.Element{
			Name: xml.Name{Space: msg.Name.Space, Local: pt.Name},
			Type: foundType,
		})
	}
	// build a complex type holding the message
	t := &xsd.ComplexType{
		Doc:        "",
		Name:       msg.Name,
		Base:       xsd.AnyType,
		TopLevel:   true,
		Elements:   elements,
		Attributes: []xsd.Attribute{},
	}
	// point the message to a single part which is the newly created complex type
	msg.Parts = []wsdl.Part{{
		Name:    msg.Name.Local,
		Type:    msg.Name,
		Element: msg.Name,
	}}
	p.wsdlschema.Types[t.Name] = t
	return msg
}

func (p *printer) genAST() error {
	p.addHelpers()
	for _, port := range p.wsdl.Ports {
		if err := p.port(port); err != nil {
			return err
		}
	}
	return nil
}

func (p *printer) port(port wsdl.Port) error {
	for _, operation := range port.Operations {
		if err := p.operation(port, operation); err != nil {
			return err
		}
	}
	return nil
}

func (p *printer) operation(port wsdl.Port, op wsdl.Operation) error {
	input, ok := p.wsdl.Message[op.Input]
	if !ok {
		return fmt.Errorf("unknown input message type %s", op.Input.Local)
	}
	output, ok := p.wsdl.Message[op.Output]
	if !ok {
		return fmt.Errorf("unknown output message type %s", op.Output.Local)
	}

	params, err := p.opArgs(port.Address, port.Method, op, input, output)
	if err != nil {
		return err
	}

	if params.InputType != "" {
		decls, err := gen.Snippets(params, `
				type {{.InputType}} struct {
				{{ range .InputFields -}}
					{{.Name}} {{.PublicType}}
				{{ end -}}
				}`,
		)
		if err != nil {
			return err
		}
		p.decl = append(p.decl, decls...)
	}
	if params.ReturnType != "" {
		decls, err := gen.Snippets(params, `
				type {{.ReturnType}} struct {
				{{ range .ReturnFields -}}
					{{.Name}} {{.Type}}
				{{ end -}}
				}`,
		)
		if err != nil {
			return err
		}
		p.decl = append(p.decl, decls...)
	}
	args := append([]string{"ctx context.Context"}, params.input...)
	fn := gen.Func(p.xsdgen.NameOf(op.Name)).
		Comment(op.Doc).
		Receiver("c *Client").
		Args(args...).
		BodyTmpl(`
		{{ if .DocumentStyle -}}
			parameters :=[]any{
			{{- range .InputFields }}
			 &{{.InputArg}},
			{{ end }}
		    }
		{{ else -}}
		parameters := struct {
			XMLName struct{} `+"`xml:\"{{.InputName.Space}} {{.InputName.Local}}\"`"+`
			{{ range .InputFields -}}
			{{.Name}} {{.Type}} `+"`"+`xml:"{{.XMLName.Space}} {{.XMLName.Local}}"`+"`"+`
			{{ end -}}
		}{ {{- range .InputFields }} {{.Name}} : {{.InputArg}},	{{ end }} }
		{{ end -}}

		{{ if .OutputFields -}}
		output := struct{
			XMLName struct{} `+"`xml:\"{{.OutputName.Space}} {{.OutputName.Local}}\"`"+`
			{{ range .OutputFields -}}
			{{.Name}} {{.Type}} `+"`"+`xml:"{{.XMLName.Space}} {{.XMLName.Local}}"`+"`"+`
			{{ end -}}
		}{}
		{{ end -}}
		response := []any{
			{{ if .DocumentStyle -}}
			{{ range .OutputFields }} 
			&output.{{.Name}} ,
			{{ end }}
			{{ else if .OutputFields -}}
			&output ,
			{{ end -}}
		}
			
			err := c.SOAP.Do(ctx, {{.SOAPAction|printf "%q"}}, &parameters, response)
			
			{{ if .OutputFields -}}
			return {{ range $index , $element := .OutputFields }} output.{{$element.Name}} , {{ end }} err
			{{- else -}}
			return err
			{{- end -}}
		`, params).
		Returns(params.output...)
	if decl, err := fn.Decl(); err != nil {
		return err
	} else {
		p.decl = append(p.decl, decl)
	}
	return nil
}

// The xsdgen package generates private types for some builtin
// types. These types should be hidden from the user and converted
// on the fly.
func exposeType(typ string) string {
	switch typ {
	case "xsdDate", "xsdTime", "xsdDateTime", "gDay",
		"gMonth", "gMonthDay", "gYear", "gYearMonth":
		return "time.Time"
	case "hexBinary", "base64Binary":
		return "[]byte"
	case "idrefs", "nmtokens", "notation", "entities":
		return "[]string"
	}
	return typ
}

func (p *printer) getPartType(part wsdl.Part) (string, error) {
	if part.Type.Local != "" {
		return p.code.NameOf(part.Type), nil
	}
	if part.Element.Local != "" {
		doc, ok := p.code.DocType(part.Element.Space)
		if !ok {
			return "", fmt.Errorf("part %s: could not lookup element %v",
				part.Name, part.Element)
		}
		for _, el := range doc.Elements {
			if el.Name == part.Element {
				return p.code.NameOf(xsd.XMLName(el.Type)), nil
			}
		}
	}
	return "", fmt.Errorf("part %s has no element or type", part.Name)
}

func (p *printer) opArgs(addr, method string, op wsdl.Operation, input, output wsdl.Message) (opArgs, error) {
	var args opArgs
	args.Address = addr
	args.Method = method
	args.SOAPAction = op.SOAPAction
	args.DocumentStyle = op.DocumentStyle
	args.MsgName = op.Name
	args.InputName = xml.Name{Local: input.Name.Local, Space: input.Name.Space}
	for _, part := range input.Parts {
		typ, err := p.getPartType(part)
		if err != nil {
			return args, err
		}
		inputType := exposeType(typ)
		vname := gen.Sanitize(part.Name)
		if vname == typ {
			vname += "_"
		}
		args.input = append(args.input, vname+" "+inputType)
		args.InputFields = append(args.InputFields, field{
			Name:       cases.Title(language.Und, cases.NoLower).String(part.Name),
			Type:       typ,
			PublicType: exposeType(typ),
			XMLName:    xml.Name{Space: p.wsdl.TargetNS, Local: part.Name},
			InputArg:   vname,
		})
	}
	if len(args.input) > p.maxArgs {
		args.InputType = cases.Title(language.Und, cases.NoLower).String(args.InputName.Local)
		args.input = []string{"v " + args.InputType}
		for i, v := range input.Parts {
			args.InputFields[i].InputArg = "v." + cases.Title(language.Und, cases.NoLower).String(v.Name)
		}
	}
	args.OutputName = xml.Name{Local: output.Name.Local, Space: output.Name.Space}
	for _, part := range output.Parts {
		typ, err := p.getPartType(part)
		if err != nil {
			return args, err
		}
		outputType := exposeType(typ)
		args.output = append(args.output, outputType)
		args.OutputFields = append(args.OutputFields, field{
			Name:    cases.Title(language.Und, cases.NoLower).String(part.Name),
			Type:    typ,
			XMLName: xml.Name{Space: p.wsdl.TargetNS, Local: part.Name},
		})
	}
	if len(args.output) > p.maxReturns {
		args.ReturnType = cases.Title(language.Und, cases.NoLower).String(args.OutputName.Local)
		args.output = []string{args.ReturnType}
		args.ReturnFields = make([]field, len(args.OutputFields))
		for i, v := range args.OutputFields {
			args.ReturnFields[i] = field{
				Name:     v.Name,
				Type:     exposeType(v.Type),
				InputArg: v.Name,
			}
		}
		//make a "virtual output field consisting of the return type
		args.OutputFields = []field{{Name: args.ReturnType, Type: args.ReturnType}}
	}
	// NOTE(droyo) if we decide to name our return values,
	// we have to change this too.
	args.output = append(args.output, "error")

	return args, nil
}
