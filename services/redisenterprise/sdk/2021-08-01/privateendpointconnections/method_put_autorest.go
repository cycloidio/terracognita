package privateendpointconnections

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

type PutResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
}

// Put ...
func (c PrivateEndpointConnectionsClient) Put(ctx context.Context, id PrivateEndpointConnectionId, input PrivateEndpointConnection) (result PutResponse, err error) {
	req, err := c.preparerForPut(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateendpointconnections.PrivateEndpointConnectionsClient", "Put", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForPut(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateendpointconnections.PrivateEndpointConnectionsClient", "Put", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// PutThenPoll performs Put then polls until it's completed
func (c PrivateEndpointConnectionsClient) PutThenPoll(ctx context.Context, id PrivateEndpointConnectionId, input PrivateEndpointConnection) error {
	result, err := c.Put(ctx, id, input)
	if err != nil {
		return fmt.Errorf("performing Put: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after Put: %+v", err)
	}

	return nil
}

// preparerForPut prepares the Put request.
func (c PrivateEndpointConnectionsClient) preparerForPut(ctx context.Context, id PrivateEndpointConnectionId, input PrivateEndpointConnection) (*http.Request, error) {
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

// senderForPut sends the Put request. The method will close the
// http.Response Body if it receives an error.
func (c PrivateEndpointConnectionsClient) senderForPut(ctx context.Context, req *http.Request) (future PutResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}
	future.Poller, err = polling.NewLongRunningPollerFromResponse(ctx, resp, c.Client)
	return
}
