package chemspell

import (
	"context"
	"encoding/xml"
	"fmt"
	"testing"
)

type SOAPmock struct{ t *testing.T }

func (s *SOAPmock) Do(ctx context.Context, action string, request any, response any) error {
	type Envelope struct {
		Body struct {
			Content []any
		}
	}
	reqEnv := new(Envelope)
	if ra, ok := request.([]any); ok {
		reqEnv.Body.Content = ra //normal case where request is []any
	} else {
		reqEnv.Body.Content = []any{request} //compatibility case where request is directly the struct
	}
	respEnv := new(Envelope)
	if ra, ok := response.([]any); ok {
		respEnv.Body.Content = ra // normal case where response is []any
	} else {
		respEnv.Body.Content = []any{response} //compatibility case where response is directly the struct
	}

	byte, err := xml.MarshalIndent(reqEnv, "", " ")
	if err != nil {
		s.t.Error(err)
	}
	if string(byte) != `<Envelope>
 <Body>
  <getSugListRequest xmlns="http://chemspell.nlm.nih.gov/axis/SpellAid.jws/axis/SpellAid.jws">
   <name xmlns="http://chemspell.nlm.nih.gov/axis/SpellAid.jws/axis/SpellAid.jws">foo</name>
   <src xmlns="http://chemspell.nlm.nih.gov/axis/SpellAid.jws/axis/SpellAid.jws">All databases</src>
  </getSugListRequest>
 </Body>
</Envelope>` {
		s.t.Error("incorrect marshaled envelope")
	}
	r := GetSugListResponse{
		Return: "bar",
	}
	rbyte, err := xml.Marshal(r)

	if re, ok := response.([]any); ok {
		for i := range re {
			if err := xml.Unmarshal(rbyte, re[i]); err == nil {
				return nil
			}
		}
	} else {
		if err := xml.Unmarshal(rbyte, response); err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to unmarshal response")
}

func TestNDFDGen(t *testing.T) {

	client := NewClient()
	client.SOAP = &SOAPmock{t}

	s, err := client.GetSugList(context.TODO(), GetSugListRequest{Name: "foo", Src: "All databases"})
	if err != nil {
		t.Error(err)
	}
	if s.Return != "bar" {
		t.Error("unexpected response")
	}
}
