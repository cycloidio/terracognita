package redisenterprise

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

type DatabasesForceUnlinkResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
}

// DatabasesForceUnlink ...
func (c RedisEnterpriseClient) DatabasesForceUnlink(ctx context.Context, id DatabaseId, input ForceUnlinkParameters) (result DatabasesForceUnlinkResponse, err error) {
	req, err := c.preparerForDatabasesForceUnlink(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "redisenterprise.RedisEnterpriseClient", "DatabasesForceUnlink", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForDatabasesForceUnlink(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "redisenterprise.RedisEnterpriseClient", "DatabasesForceUnlink", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// DatabasesForceUnlinkThenPoll performs DatabasesForceUnlink then polls until it's completed
func (c RedisEnterpriseClient) DatabasesForceUnlinkThenPoll(ctx context.Context, id DatabaseId, input ForceUnlinkParameters) error {
	result, err := c.DatabasesForceUnlink(ctx, id, input)
	if err != nil {
		return fmt.Errorf("performing DatabasesForceUnlink: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after DatabasesForceUnlink: %+v", err)
	}

	return nil
}

// preparerForDatabasesForceUnlink prepares the DatabasesForceUnlink request.
func (c RedisEnterpriseClient) preparerForDatabasesForceUnlink(ctx context.Context, id DatabaseId, input ForceUnlinkParameters) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/forceUnlink", id.ID())),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// senderForDatabasesForceUnlink sends the DatabasesForceUnlink request. The method will close the
// http.Response Body if it receives an error.
func (c RedisEnterpriseClient) senderForDatabasesForceUnlink(ctx context.Context, req *http.Request) (future DatabasesForceUnlinkResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}
	future.Poller, err = polling.NewLongRunningPollerFromResponse(ctx, resp, c.Client)
	return
}
