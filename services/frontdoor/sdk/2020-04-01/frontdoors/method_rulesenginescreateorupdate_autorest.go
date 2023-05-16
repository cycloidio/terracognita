package frontdoors

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/polling"
)

type RulesEnginesCreateOrUpdateResponse struct {
	Poller       polling.LongRunningPoller
	HttpResponse *http.Response
}

// RulesEnginesCreateOrUpdate ...
func (c FrontDoorsClient) RulesEnginesCreateOrUpdate(ctx context.Context, id RulesEngineId, input RulesEngine) (result RulesEnginesCreateOrUpdateResponse, err error) {
	req, err := c.preparerForRulesEnginesCreateOrUpdate(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "frontdoors.FrontDoorsClient", "RulesEnginesCreateOrUpdate", nil, "Failure preparing request")
		return
	}

	result, err = c.senderForRulesEnginesCreateOrUpdate(ctx, req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "frontdoors.FrontDoorsClient", "RulesEnginesCreateOrUpdate", result.HttpResponse, "Failure sending request")
		return
	}

	return
}

// RulesEnginesCreateOrUpdateThenPoll performs RulesEnginesCreateOrUpdate then polls until it's completed
func (c FrontDoorsClient) RulesEnginesCreateOrUpdateThenPoll(ctx context.Context, id RulesEngineId, input RulesEngine) error {
	result, err := c.RulesEnginesCreateOrUpdate(ctx, id, input)
	if err != nil {
		return fmt.Errorf("performing RulesEnginesCreateOrUpdate: %+v", err)
	}

	if err := result.Poller.PollUntilDone(); err != nil {
		return fmt.Errorf("polling after RulesEnginesCreateOrUpdate: %+v", err)
	}

	return nil
}

// preparerForRulesEnginesCreateOrUpdate prepares the RulesEnginesCreateOrUpdate request.
func (c FrontDoorsClient) preparerForRulesEnginesCreateOrUpdate(ctx context.Context, id RulesEngineId, input RulesEngine) (*http.Request, error) {
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

// senderForRulesEnginesCreateOrUpdate sends the RulesEnginesCreateOrUpdate request. The method will close the
// http.Response Body if it receives an error.
func (c FrontDoorsClient) senderForRulesEnginesCreateOrUpdate(ctx context.Context, req *http.Request) (future RulesEnginesCreateOrUpdateResponse, err error) {
	var resp *http.Response
	resp, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		return
	}
	future.Poller, err = polling.NewLongRunningPollerFromResponse(ctx, resp, c.Client)
	return
}
