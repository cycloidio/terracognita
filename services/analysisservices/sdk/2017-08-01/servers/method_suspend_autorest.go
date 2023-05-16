package servers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

type SuspendResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
}

// Suspend ...
func (c ServersClient) Suspend(ctx context.Context, id ServerId) (result SuspendResponse, err error) {
	req, err := c.preparerForSuspend(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "servers.ServersClient", "Suspend", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForSuspend(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "servers.ServersClient", "Suspend", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// SuspendThenPoll performs Suspend then polls until it's completed
func (c ServersClient) SuspendThenPoll(ctx context.Context, id ServerId) error {
	result, err := c.Suspend(ctx, id)
	if err != nil {
		return fmt.Errorf("performing Suspend: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after Suspend: %+v", err)
	}

	return nil
}

// preparerForSuspend prepares the Suspend request.
func (c ServersClient) preparerForSuspend(ctx context.Context, id ServerId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/suspend", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// senderForSuspend sends the Suspend request. The method will close the
// http.Response Body if it receives an error.
func (c ServersClient) senderForSuspend(ctx context.Context, req *http.Request) (future SuspendResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}
	future.Poller, err = polling.NewLongRunningPollerFromResponse(ctx, resp, c.Client)
	return
}
