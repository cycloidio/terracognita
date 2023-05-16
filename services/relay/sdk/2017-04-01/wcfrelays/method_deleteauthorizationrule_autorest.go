package wcfrelays

import (
	"context"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type DeleteAuthorizationRuleResponse struct {
	HttpResponse *http.Response
}

// DeleteAuthorizationRule ...
func (c WCFRelaysClient) DeleteAuthorizationRule(ctx context.Context, id WcfRelayAuthorizationRuleId) (result DeleteAuthorizationRuleResponse, err error) {
	req, err := c.preparerForDeleteAuthorizationRule(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "wcfrelays.WCFRelaysClient", "DeleteAuthorizationRule", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "wcfrelays.WCFRelaysClient", "DeleteAuthorizationRule", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForDeleteAuthorizationRule(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "wcfrelays.WCFRelaysClient", "DeleteAuthorizationRule", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForDeleteAuthorizationRule prepares the DeleteAuthorizationRule request.
func (c WCFRelaysClient) preparerForDeleteAuthorizationRule(ctx context.Context, id WcfRelayAuthorizationRuleId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsDelete(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(id.ID()),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForDeleteAuthorizationRule handles the response to the DeleteAuthorizationRule request. The method always
// closes the http.Response Body.
func (c WCFRelaysClient) responderForDeleteAuthorizationRule(resp *http.Response) (result DeleteAuthorizationRuleResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusNoContent, http.StatusOK),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
