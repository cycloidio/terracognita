package capacities

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type CheckNameAvailabilityResponse struct {
	HttpResponse *http.Response
	Model        *CheckCapacityNameAvailabilityResult
}

// CheckNameAvailability ...
func (c CapacitiesClient) CheckNameAvailability(ctx context.Context, id LocationId, input CheckCapacityNameAvailabilityParameters) (result CheckNameAvailabilityResponse, err error) {
	req, err := c.preparerForCheckNameAvailability(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "capacities.CapacitiesClient", "CheckNameAvailability", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "capacities.CapacitiesClient", "CheckNameAvailability", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForCheckNameAvailability(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "capacities.CapacitiesClient", "CheckNameAvailability", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForCheckNameAvailability prepares the CheckNameAvailability request.
func (c CapacitiesClient) preparerForCheckNameAvailability(ctx context.Context, id LocationId, input CheckCapacityNameAvailabilityParameters) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/checkNameAvailability", id.ID())),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForCheckNameAvailability handles the response to the CheckNameAvailability request. The method always
// closes the http.Response Body.
func (c CapacitiesClient) responderForCheckNameAvailability(resp *http.Response) (result CheckNameAvailabilityResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
