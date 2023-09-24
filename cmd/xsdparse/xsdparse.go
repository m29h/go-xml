package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/m29h/go-xml/xmltree"
	"github.com/m29h/go-xml/xsd"
)

var (
	TargetNS = flag.String("ns", "", "Namespace of schema to print")
)

func main() {
	log.SetFlags(0)
	// Usage is a replacement usage function for the flags package.
	flag.Usage = func() {
		prog := os.Args[0]
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", prog)
		fmt.Fprintf(os.Stderr, "\t%s [flags] file(s)... # files must have xsd/wsdl schema content\n", prog)
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	docs := make([][]byte, 0, flag.NArg())

	for _, filename := range flag.Args() {
		if data, err := os.ReadFile(filename); err != nil {
			log.Fatal(err)
		} else {
			docs = append(docs, data)
		}
	}

	filterSchema := make(map[string]struct{})
	for _, doc := range xsd.StandardSchema {
		root, err := xmltree.Parse(doc)
		if err != nil {
			// should never happen
			panic(err)
		}
		filterSchema[root.Attr("", "targetNamespace")] = struct{}{}
	}

	norm, err := xsd.Normalize(docs...)
	if err != nil {
		log.Fatal(err)
	}

	selected := make([]*xmltree.Element, 0, len(norm))
	for _, root := range norm {
		tns := root.Attr("", "targetNamespace")
		if *TargetNS != "" && *TargetNS == tns {
			selected = append(selected, root)
		} else if _, ok := filterSchema[tns]; !ok {
			selected = append(selected, root)
		}
	}

	for _, root := range selected {
		fmt.Printf("%s\n", xmltree.MarshalIndent(root, "", "  "))
	}
}
