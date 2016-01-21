package soap

import (
	"encoding/xml"
	"errors"
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

func (s *Server) serveSOAP(requestEnvelopeBytes []byte, soapAction string) (responseEnvelopeBytes []byte, err error) {
	messageType := "checkVatRequest"
	soapAction = "operationCheckVat"
	actionHandlers, ok := s.handlers[soapAction]
	if !ok {
		err = errors.New("could not find handlers for action: \"" + soapAction + "\"")
		return
	}
	handler, ok := actionHandlers[messageType]
	if !ok {
		err = errors.New("no handler for message type: " + messageType)
		return
	}
	request := handler.requestFactory()
	// parse from envelope.body.content into request
	response, err := handler.handler(request)
	responseEnvelope := &SOAPEnvelope{
		Body: SOAPBody{},
	}
	if err != nil {
		// soap fault party time
		responseEnvelope.Body.Fault = &SOAPFault{
			String: err.Error(),
		}
	} else {
		responseEnvelope.Body.Content = response
	}
	// marshal responseEnvelope
	return
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
		operationHandlers, ok := s.handlers[""]
		if !ok {
			// soap fault
			return
		}
		operationHandler, ok := operationHandlers["checkVatRequest"]
		if !ok {
			// soap fault
			return
		}
		request := operationHandler.requestFactory()

		envelope := &SOAPEnvelope{
			Body: SOAPBody{
				Content: request,
			},
		}
		errUnm := xml.Unmarshal(requestBytes, envelope)

		fmt.Println("Request Body", envelope, "Error:", errUnm)

		response, errResp := operationHandler.handler(envelope.Body.Content)

		if errResp != nil {
			// soap fault
			return
		}

		responseEnvelope := &SOAPEnvelope{
			Body: SOAPBody{
				Content: response,
			},
		}

		xmlBytes, err := xml.MarshalIndent(responseEnvelope, "", "	")

		if err != nil {
			// soap fault
			return
		}
		w.Write(xmlBytes)
	default:
		// this will be a soap fault !?
		w.Write([]byte("this is a soap service - you have to POST soap requests\n"))
		w.Write([]byte("invalid method: " + r.Method))
	}
}

func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}
