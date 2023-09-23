package forecast

import (
	"context"
	"encoding/xml"
	"fmt"
	"testing"
	"time"
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

	reqByte, err := xml.MarshalIndent(reqEnv, "", " ")
	if err != nil {
		s.t.Error(err)
	}
	want := `<Envelope>
 <Body>
  <NDFDgenRequest xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">
   <latitude xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">27</latitude>
   <longitude xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">-100</longitude>
   <product xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">time-series</product>
   <Unit xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">m</Unit>
   <weatherParameters xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">
    <maxt xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</maxt>
    <mint xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</mint>
    <temp xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</temp>
    <dew xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</dew>
    <pop12 xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</pop12>
    <qpf xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</qpf>
    <sky xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">true</sky>
    <snow xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</snow>
    <wspd xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</wspd>
    <wdir xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</wdir>
    <wx xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</wx>
    <waveh xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</waveh>
    <icons xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</icons>
    <rh xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</rh>
    <appt xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</appt>
    <incw34 xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</incw34>
    <incw50 xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</incw50>
    <incw64 xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</incw64>
    <cumw34 xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</cumw34>
    <cumw50 xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</cumw50>
    <cumw64 xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</cumw64>
    <critfireo xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</critfireo>
    <dryfireo xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</dryfireo>
    <conhazo xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</conhazo>
    <ptornado xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</ptornado>
    <phail xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</phail>
    <ptstmwinds xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</ptstmwinds>
    <pxtornado xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</pxtornado>
    <pxhail xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</pxhail>
    <pxtstmwinds xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</pxtstmwinds>
    <ptotsvrtstm xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</ptotsvrtstm>
    <pxtotsvrtstm xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</pxtotsvrtstm>
    <tmpabv14d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</tmpabv14d>
    <tmpblw14d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</tmpblw14d>
    <tmpabv30d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</tmpabv30d>
    <tmpblw30d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</tmpblw30d>
    <tmpabv90d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</tmpabv90d>
    <tmpblw90d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</tmpblw90d>
    <prcpabv14d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</prcpabv14d>
    <prcpblw14d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</prcpblw14d>
    <prcpabv30d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</prcpabv30d>
    <prcpblw30d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</prcpblw30d>
    <prcpabv90d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</prcpabv90d>
    <prcpblw90d xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</prcpblw90d>
    <precipa_r xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</precipa_r>
    <sky_r xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</sky_r>
    <td_r xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</td_r>
    <temp_r xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</temp_r>
    <wdir_r xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</wdir_r>
    <wspd_r xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</wspd_r>
    <wwa xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</wwa>
    <wgust xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</wgust>
    <iceaccum xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</iceaccum>
    <maxrh xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</maxrh>
    <minrh xmlns="http://graphical.weather.gov/xml/DWMLgen/schema/DWML.xsd">false</minrh>
   </weatherParameters>
   <startTime xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">2022-01-01T12:00:00.234567Z</startTime>
   <endTime xmlns="http://graphical.weather.gov/xml/DWMLgen/wsdl/ndfdXML.wsdl">2023-01-01T12:00:00.123456Z</endTime>
  </NDFDgenRequest>
 </Body>
</Envelope>`
	if string(reqByte) != want {
		fmt.Printf("got:\n%s\nwant:\n%s\n", string(reqByte), want)
		s.t.Error("incorrect marshaled envelope")
	}
	r := NDFDgenResponse{
		DwmlOut: "bar",
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

	s, err := client.NDFDgen(context.TODO(), NDFDgenRequest{
		EndTime:   time.Date(2023, 01, 01, 12, 00, 00, 123456789, time.UTC),
		StartTime: time.Date(2022, 01, 01, 12, 00, 00, 234567891, time.UTC),
		Unit:      "m",
		Product:   "time-series",
		Latitude:  27,
		Longitude: -100,
		WeatherParameters: WeatherParameters{
			Sky: true,
		},
	})
	if err != nil {
		t.Error(err)
	}
	if s.DwmlOut != "bar" {
		t.Errorf("received `%s`, expected `bar`", s.DwmlOut)
	}
	t.Log(s)
}
