package soap

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ########### REMOVE THIS ###############
type CheckVatRequest struct {
	XMLName xml.Name `xml:"urn:ec.europa.eu:taxud:vies:services:checkVat:types checkVat"`

	CountryCode string `xml:"countryCode,omitempty"`
	VatNumber   string `xml:"vatNumber,omitempty"`
}

type CheckVatResponse struct {
	XMLName xml.Name `xml:"urn:ec.europa.eu:taxud:vies:services:checkVat:types checkVatResponse"`

	CountryCode string `xml:"countryCode,omitempty"`
	VatNumber   string `xml:"vatNumber,omitempty"`
	RequestDate string `xml:"requestDate,omitempty"`
	Valid       bool   `xml:"valid,omitempty"`
	Name        string `xml:"name,omitempty"`
	Address     string `xml:"address,omitempty"`
}

// ########### REMOVE THIS END ###############

type OperationHandlerFunc func(request interface{}) (response interface{}, err error)
type RequestFactoryFunc func() interface{}

type operationHandler struct {
	requestFactory RequestFactoryFunc
	handler        OperationHandlerFunc
}

type Server struct {
	handlers map[string]map[string]*operationHandler
}

func NewServer() *Server {
	s := &Server{
		handlers: make(map[string]map[string]*operationHandler),
	}
	return s
}

// HandleOperation register to handle an operation
func (s *Server) HandleOperation(action string, messageType string, requestFactory RequestFactoryFunc, operationHandlerFunc OperationHandlerFunc) {
	_, ok := s.handlers[action]
	if !ok {
		s.handlers[action] = make(map[string]*operationHandler) // is this good??
	}

	s.handlers[action][messageType] = &operationHandler{
		handler:        operationHandlerFunc,
		requestFactory: requestFactory,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		w.Write([]byte("that actually could be a soap request"))
		requestBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
		}

		// // Retrieve the appropriate handler
		// @todo be more careful retrieving the operation handler
		soapaction := r.Header.Get("SOAPAction")
		// TODO get parse soap action from request if it is not available in header

		fmt.Println("SOAPACTION: ", soapaction)
		operationHandlers, ok := s.handlers[soapaction]
		if !ok {
			w.Write(newSoapFault("Code", "no action found", "actor", "detail"))
			return
		}
		operationHandler, ok := operationHandlers["checkVatRequest"] // TODO replace hard coded request
		if !ok {
			w.Write(newSoapFault("Code", "no handler found", "actor", "detail"))
			return
		}
		request := operationHandler.requestFactory()

		envelope := &SOAPEnvelope{
			Body: SOAPBody{
				Content: request,
			},
		}
		errUnm := xml.Unmarshal(requestBytes, envelope)

		fmt.Println("Request", request, "Request Body", envelope, "Error:", errUnm)

		response, errResp := operationHandler.handler(envelope.Body.Content)
		if errResp != nil {
			// soap fault
			w.Write(newSoapFault("Code", errResp.Error(), "actor", "detail"))
			return
		}

		responseEnvelope := &SOAPEnvelope{
			Body: SOAPBody{
				Content: response,
			},
		}

		xmlBytes, err := xml.MarshalIndent(responseEnvelope, "", "	")

		if err != nil {
			w.Write(newSoapFault("Code", err.Error(), "actor", "detail"))
			return
		}
		w.Write(xmlBytes)
	default:
		// this will be a soap fault !?
		w.Write(newSoapFault("Code", "Invalid Request. Use POST method! ", "actor", "detail"))
	}
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

func newSoapFault(faultCode string, faultString string, faultActor string, detail string) []byte {
	soapfault := &SOAPFault{
		Code:   faultCode,
		String: faultString,
		Actor:  faultActor,
		Detail: detail,
	}
	xmlBytes, _ := xml.MarshalIndent(soapfault, "", "	")
	return xmlBytes
}
