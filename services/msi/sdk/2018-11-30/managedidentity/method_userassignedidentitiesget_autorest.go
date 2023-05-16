package managedidentity

import (
	"context"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type UserAssignedIdentitiesGetResponse struct {
	HttpResponse *http.Response
	Model        *Identity
}

// UserAssignedIdentitiesGet ...
func (c ManagedIdentityClient) UserAssignedIdentitiesGet(ctx context.Context, id UserAssignedIdentitiesId) (result UserAssignedIdentitiesGetResponse, err error) {
	req, err := c.preparerForUserAssignedIdentitiesGet(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "managedidentity.ManagedIdentityClient", "UserAssignedIdentitiesGet", nil, "Failure preparing request")
		return
	}

	result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "managedidentity.ManagedIdentityClient", "UserAssignedIdentitiesGet", result.HttpResponse, "Failure sending request")
		return
	}

	result, err = c.responderForUserAssignedIdentitiesGet(result.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "managedidentity.ManagedIdentityClient", "UserAssignedIdentitiesGet", result.HttpResponse, "Failure responding to request")
		return
	}

	return
}

// preparerForUserAssignedIdentitiesGet prepares the UserAssignedIdentitiesGet request.
func (c ManagedIdentityClient) preparerForUserAssignedIdentitiesGet(ctx context.Context, id UserAssignedIdentitiesId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(id.ID()),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForUserAssignedIdentitiesGet handles the response to the UserAssignedIdentitiesGet request. The method always
// closes the http.Response Body.
func (c ManagedIdentityClient) responderForUserAssignedIdentitiesGet(resp *http.Response) (result UserAssignedIdentitiesGetResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result.Model),
		autorest.ByClosing())
	result.HttpResponse = resp
	return
}
