// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/netlify/open-api/v2/go/models"
)

// DeleteDNSZoneReader is a Reader for the DeleteDNSZone structure.
type DeleteDNSZoneReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteDNSZoneReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewDeleteDNSZoneNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewDeleteDNSZoneDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDeleteDNSZoneNoContent creates a DeleteDNSZoneNoContent with default headers values
func NewDeleteDNSZoneNoContent() *DeleteDNSZoneNoContent {
	return &DeleteDNSZoneNoContent{}
}

/*
DeleteDNSZoneNoContent handles this case with default header values.

delete a single DNS zone
*/
type DeleteDNSZoneNoContent struct {
}

func (o *DeleteDNSZoneNoContent) Error() string {
	return fmt.Sprintf("[DELETE /dns_zones/{zone_id}][%d] deleteDnsZoneNoContent ", 204)
}

func (o *DeleteDNSZoneNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteDNSZoneDefault creates a DeleteDNSZoneDefault with default headers values
func NewDeleteDNSZoneDefault(code int) *DeleteDNSZoneDefault {
	return &DeleteDNSZoneDefault{
		_statusCode: code,
	}
}

/*
DeleteDNSZoneDefault handles this case with default header values.

error
*/
type DeleteDNSZoneDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the delete Dns zone default response
func (o *DeleteDNSZoneDefault) Code() int {
	return o._statusCode
}

func (o *DeleteDNSZoneDefault) Error() string {
	return fmt.Sprintf("[DELETE /dns_zones/{zone_id}][%d] deleteDnsZone default  %+v", o._statusCode, o.Payload)
}

func (o *DeleteDNSZoneDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *DeleteDNSZoneDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}