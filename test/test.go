package main

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/foomo/soap"
)

type FooRequest struct {
	Foo string
}

type FooResponse struct {
	Bar string
}

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
	Valid       bool   `xml:"valid"`
	Name        string `xml:"name,omitempty"`
	Address     string `xml:"address,omitempty"`
}

func RunServer() {
	soapServer := soap.NewServer()

	// soapServer.HandleOperation(
	// 	"operationFoo",
	// 	"FooRequest",
	// 	func() interface{} {
	// 		return &FooRequest{}
	// 	},
	// 	func(request interface{}) (response interface{}, err error) {
	// 		fooRequest := request.(*FooRequest)
	// 		fooResponse := &FooResponse{
	// 			Bar: "Hello " + fooRequest.Foo,
	// 		}
	// 		response = fooResponse
	// 		return
	// 	},
	// )

	// Create handler for checkVat
	soapServer.HandleOperation(
		"operationCheckVat",
		"checkVatRequest",
		func() interface{} {
			r := &CheckVatRequest{}
			log.Println("creating new vat request", r)
			return r
		},
		func(request interface{}) (response interface{}, err error) {
			log.Println("handling request", request)

			checkVatRequest := request.(*CheckVatRequest)
			checkVatResponse := &CheckVatResponse{
				CountryCode: checkVatRequest.CountryCode,
				VatNumber:   "lalala",
				Valid:       false,
			}
			response = checkVatResponse
			return
		},
	)

	err := soapServer.ListenAndServe(":8080")
	fmt.Println(err)
}

func main() {
	RunServer()
}
