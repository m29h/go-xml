package wsdlgen

import "aqwari.net/xml/internal/gen"

// One of the goals of this package is that generated code
// has no external dependencies, only the Go standard
// library. That means we have to bundle any static
// "helper" functions along with the generated code. We
// are playing a balancing game here; the larger the static
// code base grows, the weaker the argument against external
// dependencies becomes.
var helpers string = `
type SOAPdoer interface {
	Do(ctx context.Context, action string, request any, response any) error
}

type Client struct {
	SOAP  SOAPdoer
}
`

func (p *printer) addHelpers() {
	decls, err := gen.Declarations(helpers)
	if err != nil {
		// code does not change at runtime, so
		// this should never happen
		panic(err)
	}
	p.decl = append(p.decl, decls...)
}
