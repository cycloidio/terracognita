package privateendpointconnections

import (
	"context"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type CreateResponse struct {
	HttpResponse *http.Response
	Model        *PrivateEndpointConnection
}

// Create ...
func (c PrivateEndpointConnectionsClient) Create(ctx context.Context, id PrivateEndpointConnectionId, input PrivateEndpointConnection) (result CreateResponse, err error) {
	req, err := c.preparerForCreate(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateendpointconnections.PrivateEndpointConnectionsClient", "Create", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateendpointconnections.PrivateEndpointConnectionsClient", "Create", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForCreate(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateendpointconnections.PrivateEndpointConnectionsClient", "Create", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForCreate prepares the Create request.
func (c PrivateEndpointConnectionsClient) preparerForCreate(ctx context.Context, id PrivateEndpointConnectionId, input PrivateEndpointConnection) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPut(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(id.ID()),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForCreate handles the response to the Create request. The method always
// closes the http.Response Body.
func (c PrivateEndpointConnectionsClient) responderForCreate(resp *http.Response) (result CreateResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
