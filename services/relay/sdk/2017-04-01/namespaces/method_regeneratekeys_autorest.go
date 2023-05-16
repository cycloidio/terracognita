package namespaces

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type RegenerateKeysResponse struct {
	HttpResponse *http.Response
	Model        *AccessKeys
}

// RegenerateKeys ...
func (c NamespacesClient) RegenerateKeys(ctx context.Context, id AuthorizationRuleId, input RegenerateAccessKeyParameters) (result RegenerateKeysResponse, err error) {
	req, err := c.preparerForRegenerateKeys(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "namespaces.NamespacesClient", "RegenerateKeys", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "namespaces.NamespacesClient", "RegenerateKeys", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForRegenerateKeys(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "namespaces.NamespacesClient", "RegenerateKeys", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForRegenerateKeys prepares the RegenerateKeys request.
func (c NamespacesClient) preparerForRegenerateKeys(ctx context.Context, id AuthorizationRuleId, input RegenerateAccessKeyParameters) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/regenerateKeys", id.ID())),
		autorest.WithJSON(input),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForRegenerateKeys handles the response to the RegenerateKeys request. The method always
// closes the http.Response Body.
func (c NamespacesClient) responderForRegenerateKeys(resp *http.Response) (result RegenerateKeysResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
