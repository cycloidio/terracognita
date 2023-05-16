package databases

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

type ImportResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
}

// Import ...
func (c DatabasesClient) Import(ctx context.Context, id DatabaseId, input ImportClusterParameters) (result ImportResponse, err error) {
	req, err := c.preparerForImport(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "databases.DatabasesClient", "Import", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForImport(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "databases.DatabasesClient", "Import", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// ImportThenPoll performs Import then polls until it's completed
func (c DatabasesClient) ImportThenPoll(ctx context.Context, id DatabaseId, input ImportClusterParameters) error {
	result, err := c.Import(ctx, id, input)
	if err != nil {
		return fmt.Errorf("performing Import: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after Import: %+v", err)
	}

	return nil
}

// preparerForImport prepares the Import request.
func (c DatabasesClient) preparerForImport(ctx context.Context, id DatabaseId, input ImportClusterParameters) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/import", id.ID())),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// senderForImport sends the Import request. The method will close the
// http.Response Body if it receives an error.
func (c DatabasesClient) senderForImport(ctx context.Context, req *http.Request) (future ImportResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}
	future.Poller, err = polling.NewLongRunningPollerFromResponse(ctx, resp, c.Client)
	return
}
