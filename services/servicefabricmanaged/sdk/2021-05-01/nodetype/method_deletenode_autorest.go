package nodetype

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

type DeleteNodeResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
}

// DeleteNode ...
func (c NodeTypeClient) DeleteNode(ctx context.Context, id NodeTypeId, input NodeTypeActionParameters) (result DeleteNodeResponse, err error) {
	req, err := c.preparerForDeleteNode(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "DeleteNode", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForDeleteNode(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "DeleteNode", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// DeleteNodeThenPoll performs DeleteNode then polls until it's completed
func (c NodeTypeClient) DeleteNodeThenPoll(ctx context.Context, id NodeTypeId, input NodeTypeActionParameters) error {
	result, err := c.DeleteNode(ctx, id, input)
	if err != nil {
		return fmt.Errorf("performing DeleteNode: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after DeleteNode: %+v", err)
	}

	return nil
}

// preparerForDeleteNode prepares the DeleteNode request.
func (c NodeTypeClient) preparerForDeleteNode(ctx context.Context, id NodeTypeId, input NodeTypeActionParameters) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/deleteNode", id.ID())),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// senderForDeleteNode sends the DeleteNode request. The method will close the
// http.Response Body if it receives an error.
func (c NodeTypeClient) senderForDeleteNode(ctx context.Context, req *http.Request) (future DeleteNodeResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}
	future.Poller, err = polling.NewLongRunningPollerFromResponse(ctx, resp, c.Client)
	return
}
