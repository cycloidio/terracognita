package hybridconnections

import (
	"context"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type CreateOrUpdateAuthorizationRuleResponse struct {
	HttpResponse *http.Response
	Model        *AuthorizationRule
}

// CreateOrUpdateAuthorizationRule ...
func (c HybridConnectionsClient) CreateOrUpdateAuthorizationRule(ctx context.Context, id HybridConnectionAuthorizationRuleId, input AuthorizationRule) (result CreateOrUpdateAuthorizationRuleResponse, err error) {
	req, err := c.preparerForCreateOrUpdateAuthorizationRule(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "hybridconnections.HybridConnectionsClient", "CreateOrUpdateAuthorizationRule", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "hybridconnections.HybridConnectionsClient", "CreateOrUpdateAuthorizationRule", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForCreateOrUpdateAuthorizationRule(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "hybridconnections.HybridConnectionsClient", "CreateOrUpdateAuthorizationRule", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForCreateOrUpdateAuthorizationRule prepares the CreateOrUpdateAuthorizationRule request.
func (c HybridConnectionsClient) preparerForCreateOrUpdateAuthorizationRule(ctx context.Context, id HybridConnectionAuthorizationRuleId, input AuthorizationRule) (*http.Request, error) {
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

// responderForCreateOrUpdateAuthorizationRule handles the response to the CreateOrUpdateAuthorizationRule request. The method always
// closes the http.Response Body.
func (c HybridConnectionsClient) responderForCreateOrUpdateAuthorizationRule(resp *http.Response) (result CreateOrUpdateAuthorizationRuleResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
