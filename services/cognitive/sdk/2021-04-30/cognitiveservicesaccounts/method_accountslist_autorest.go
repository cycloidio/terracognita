package cognitiveservicesaccounts

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type AccountsListResponse struct {
	HttpResponse *http.Response
	Model        *[]Account

	nextLink     *string
	nextPageFunc func(ctx context.Context, nextLink string) (AccountsListResponse, error)
}

type AccountsListCompleteResult struct {
	Items []Account
}

func (r AccountsListResponse) HasMore() bool {
	return r.nextLink != nil
}

func (r AccountsListResponse) LoadMore(ctx context.Context) (resp AccountsListResponse, err error) {
	if !r.HasMore() {
		err = fmt.Errorf("no more pages returned")
		return
	}
	return r.nextPageFunc(ctx, *r.nextLink)
}

// AccountsList ...
func (c CognitiveServicesAccountsClient) AccountsList(ctx context.Context, id SubscriptionId) (resp AccountsListResponse, err error) {
	req, err := c.preparerForAccountsList(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "cognitiveservicesaccounts.CognitiveServicesAccountsClient", "AccountsList", nil, "Failure preparing request")
		return
	}

	resp.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "cognitiveservicesaccounts.CognitiveServicesAccountsClient", "AccountsList", resp.HttpResponse, "Failure sending request")
		return
	}

	resp, err = c.responderForAccountsList(resp.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "cognitiveservicesaccounts.CognitiveServicesAccountsClient", "AccountsList", resp.HttpResponse, "Failure responding to request")
		return
	}
	return
}

// AccountsListComplete retrieves all of the results into a single object
func (c CognitiveServicesAccountsClient) AccountsListComplete(ctx context.Context, id SubscriptionId) (AccountsListCompleteResult, error) {
	return c.AccountsListCompleteMatchingPredicate(ctx, id, AccountPredicate{})
}

// AccountsListCompleteMatchingPredicate retrieves all of the results and then applied the predicate
func (c CognitiveServicesAccountsClient) AccountsListCompleteMatchingPredicate(ctx context.Context, id SubscriptionId, predicate AccountPredicate) (resp AccountsListCompleteResult, err error) {
	items := make([]Account, 0)

	page, err := c.AccountsList(ctx, id)
	if err != nil {
		err = fmt.Errorf("loading the initial page: %+v", err)
		return
	}
	if page.Model != nil {
		for _, v := range *page.Model {
			if predicate.Matches(v) {
				items = append(items, v)
			}
		}
	}

	for page.HasMore() {
		page, err = page.LoadMore(ctx)
		if err != nil {
			err = fmt.Errorf("loading the next page: %+v", err)
			return
		}

		if page.Model != nil {
			for _, v := range *page.Model {
				if predicate.Matches(v) {
					items = append(items, v)
				}
			}
		}
	}

	out := AccountsListCompleteResult{
		Items: items,
	}
	return out, nil
}

// preparerForAccountsList prepares the AccountsList request.
func (c CognitiveServicesAccountsClient) preparerForAccountsList(ctx context.Context, id SubscriptionId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/providers/Microsoft.CognitiveServices/accounts", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// preparerForAccountsListWithNextLink prepares the AccountsList request with the given nextLink token.
func (c CognitiveServicesAccountsClient) preparerForAccountsListWithNextLink(ctx context.Context, nextLink string) (*http.Request, error) {
	uri, err := url.Parse(nextLink)
	if err != nil {
		return nil, fmt.Errorf("parsing nextLink %q: %+v", nextLink, err)
	}
	queryParameters := map[string]interface{}{}
	for k, v := range uri.Query() {
		if len(v) == 0 {
			continue
		}
		val := v[0]
		val = autorest.Encode("query", val)
		queryParameters[k] = val
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(uri.Path),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// responderForAccountsList handles the response to the AccountsList request. The method always
// closes the http.Response Body.
func (c CognitiveServicesAccountsClient) responderForAccountsList(resp *http.Response) (result AccountsListResponse, err error) {
	type page struct {
		Values   []Account `json:"value"`
		NextLink *string   `json:"nextLink"`
	}
	var respObj page
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&respObj),
		autorest.ByClosing())
	result.HttpResponse = resp
	result.Model = &respObj.Values
	result.nextLink = respObj.NextLink
	if respObj.NextLink != nil {
		result.nextPageFunc = func(ctx context.Context, nextLink string) (result AccountsListResponse, err error) {
			req, err := c.preparerForAccountsListWithNextLink(ctx, nextLink)
			if err != nil {
				err = autorest.NewErrorWithError(err, "cognitiveservicesaccounts.CognitiveServicesAccountsClient", "AccountsList", nil, "Failure preparing request")
				return
			}

			result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
			if err != nil {
				err = autorest.NewErrorWithError(err, "cognitiveservicesaccounts.CognitiveServicesAccountsClient", "AccountsList", result.HttpResponse, "Failure sending request")
				return
			}

			result, err = c.responderForAccountsList(result.HttpResponse)
			if err != nil {
				err = autorest.NewErrorWithError(err, "cognitiveservicesaccounts.CognitiveServicesAccountsClient", "AccountsList", result.HttpResponse, "Failure responding to request")
				return
			}

			return
		}
	}
	return
}
