// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/netlify/open-api/v2/go/models"
)

// NewCreateSplitTestParams creates a new CreateSplitTestParams object
// with the default values initialized.
func NewCreateSplitTestParams() *CreateSplitTestParams {
	var ()
	return &CreateSplitTestParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCreateSplitTestParamsWithTimeout creates a new CreateSplitTestParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCreateSplitTestParamsWithTimeout(timeout time.Duration) *CreateSplitTestParams {
	var ()
	return &CreateSplitTestParams{

		timeout: timeout,
	}
}

// NewCreateSplitTestParamsWithContext creates a new CreateSplitTestParams object
// with the default values initialized, and the ability to set a context for a request
func NewCreateSplitTestParamsWithContext(ctx context.Context) *CreateSplitTestParams {
	var ()
	return &CreateSplitTestParams{

		Context: ctx,
	}
}

// NewCreateSplitTestParamsWithHTTPClient creates a new CreateSplitTestParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCreateSplitTestParamsWithHTTPClient(client *http.Client) *CreateSplitTestParams {
	var ()
	return &CreateSplitTestParams{
		HTTPClient: client,
	}
}

/*
CreateSplitTestParams contains all the parameters to send to the API endpoint
for the create split test operation typically these are written to a http.Request
*/
type CreateSplitTestParams struct {

	/*BranchTests*/
	BranchTests *models.SplitTestSetup
	/*SiteID*/
	SiteID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the create split test params
func (o *CreateSplitTestParams) WithTimeout(timeout time.Duration) *CreateSplitTestParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create split test params
func (o *CreateSplitTestParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create split test params
func (o *CreateSplitTestParams) WithContext(ctx context.Context) *CreateSplitTestParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create split test params
func (o *CreateSplitTestParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create split test params
func (o *CreateSplitTestParams) WithHTTPClient(client *http.Client) *CreateSplitTestParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create split test params
func (o *CreateSplitTestParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBranchTests adds the branchTests to the create split test params
func (o *CreateSplitTestParams) WithBranchTests(branchTests *models.SplitTestSetup) *CreateSplitTestParams {
	o.SetBranchTests(branchTests)
	return o
}

// SetBranchTests adds the branchTests to the create split test params
func (o *CreateSplitTestParams) SetBranchTests(branchTests *models.SplitTestSetup) {
	o.BranchTests = branchTests
}

// WithSiteID adds the siteID to the create split test params
func (o *CreateSplitTestParams) WithSiteID(siteID string) *CreateSplitTestParams {
	o.SetSiteID(siteID)
	return o
}

// SetSiteID adds the siteId to the create split test params
func (o *CreateSplitTestParams) SetSiteID(siteID string) {
	o.SiteID = siteID
}

// WriteToRequest writes these params to a swagger request
func (o *CreateSplitTestParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.BranchTests != nil {
		if err := r.SetBodyParam(o.BranchTests); err != nil {
			return err
		}
	}

	// path param site_id
	if err := r.SetPathParam("site_id", o.SiteID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}