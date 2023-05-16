package servers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListSkusForExistingResponse struct {
	HttpResponse *http.Response
	Model        *SkuEnumerationForExistingResourceResult
}

// ListSkusForExisting ...
func (c ServersClient) ListSkusForExisting(ctx context.Context, id ServerId) (result ListSkusForExistingResponse, err error) {
	req, err := c.preparerForListSkusForExisting(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "servers.ServersClient", "ListSkusForExisting", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "servers.ServersClient", "ListSkusForExisting", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForListSkusForExisting(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "servers.ServersClient", "ListSkusForExisting", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForListSkusForExisting prepares the ListSkusForExisting request.
func (c ServersClient) preparerForListSkusForExisting(ctx context.Context, id ServerId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/skus", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForListSkusForExisting handles the response to the ListSkusForExisting request. The method always
// closes the http.Response Body.
func (c ServersClient) responderForListSkusForExisting(resp *http.Response) (result ListSkusForExistingResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
