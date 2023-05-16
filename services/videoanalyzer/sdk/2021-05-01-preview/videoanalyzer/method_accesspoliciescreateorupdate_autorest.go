package videoanalyzer

import (
	"context"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type AccessPoliciesCreateOrUpdateResponse struct {
	HttpResponse *http.Response
	Model        *AccessPolicyEntity
}

// AccessPoliciesCreateOrUpdate ...
func (c VideoAnalyzerClient) AccessPoliciesCreateOrUpdate(ctx context.Context, id AccessPoliciesId, input AccessPolicyEntity) (result AccessPoliciesCreateOrUpdateResponse, err error) {
	req, err := c.preparerForAccessPoliciesCreateOrUpdate(ctx, id, input)
	if err != nil {
		err = autorest.NewErrorWithError(err, "videoanalyzer.VideoAnalyzerClient", "AccessPoliciesCreateOrUpdate", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "videoanalyzer.VideoAnalyzerClient", "AccessPoliciesCreateOrUpdate", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForAccessPoliciesCreateOrUpdate(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "videoanalyzer.VideoAnalyzerClient", "AccessPoliciesCreateOrUpdate", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForAccessPoliciesCreateOrUpdate prepares the AccessPoliciesCreateOrUpdate request.
func (c VideoAnalyzerClient) preparerForAccessPoliciesCreateOrUpdate(ctx context.Context, id AccessPoliciesId, input AccessPolicyEntity) (*http.Request, error) {
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

// responderForAccessPoliciesCreateOrUpdate handles the response to the AccessPoliciesCreateOrUpdate request. The method always
// closes the http.Response Body.
func (c VideoAnalyzerClient) responderForAccessPoliciesCreateOrUpdate(resp *http.Response) (result AccessPoliciesCreateOrUpdateResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusCreated, http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
