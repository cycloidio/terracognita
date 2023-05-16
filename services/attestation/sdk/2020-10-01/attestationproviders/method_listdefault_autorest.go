package attestationproviders

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListDefaultResponse struct {
	HttpResponse *http.Response
	Model        *AttestationProviderListResult
}

// ListDefault ...
func (c AttestationProvidersClient) ListDefault(ctx context.Context, id SubscriptionId) (result ListDefaultResponse, err error) {
	req, err := c.preparerForListDefault(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "attestationproviders.AttestationProvidersClient", "ListDefault", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "attestationproviders.AttestationProvidersClient", "ListDefault", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForListDefault(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "attestationproviders.AttestationProvidersClient", "ListDefault", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForListDefault prepares the ListDefault request.
func (c AttestationProvidersClient) preparerForListDefault(ctx context.Context, id SubscriptionId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/providers/Microsoft.Attestation/defaultProviders", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForListDefault handles the response to the ListDefault request. The method always
// closes the http.Response Body.
func (c AttestationProvidersClient) responderForListDefault(resp *http.Response) (result ListDefaultResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
