package privateclouds

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListInSubscriptionResponse struct {
	HttpResponse *http.Response
	Model        *[]PrivateCloud

	nextLink     *string
	nextPageFunc func(ctx context.Context, nextLink string) (ListInSubscriptionResponse, error)
}

type ListInSubscriptionCompleteResult struct {
	Items []PrivateCloud
}

func (r ListInSubscriptionResponse) HasMore() bool {
	return r.nextLink != nil
}

func (r ListInSubscriptionResponse) LoadMore(ctx context.Context) (resp ListInSubscriptionResponse, err error) {
	if !r.HasMore() {
		err = fmt.Errorf("no more pages returned")
		return
	}
	return r.nextPageFunc(ctx, *r.nextLink)
}

// ListInSubscription ...
func (c PrivateCloudsClient) ListInSubscription(ctx context.Context, id SubscriptionId) (resp ListInSubscriptionResponse, err error) {
	req, err := c.preparerForListInSubscription(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateclouds.PrivateCloudsClient", "ListInSubscription", nil, "Failure preparing request")
		return
	}

	resp.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateclouds.PrivateCloudsClient", "ListInSubscription", resp.HttpResponse, "Failure sending request")
		return
	}

	resp, err = c.responderForListInSubscription(resp.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "privateclouds.PrivateCloudsClient", "ListInSubscription", resp.HttpResponse, "Failure responding to request")
		return
	}
	return
}

// ListInSubscriptionComplete retrieves all of the results into a single object
func (c PrivateCloudsClient) ListInSubscriptionComplete(ctx context.Context, id SubscriptionId) (ListInSubscriptionCompleteResult, error) {
	return c.ListInSubscriptionCompleteMatchingPredicate(ctx, id, PrivateCloudPredicate{})
}

// ListInSubscriptionCompleteMatchingPredicate retrieves all of the results and then applied the predicate
func (c PrivateCloudsClient) ListInSubscriptionCompleteMatchingPredicate(ctx context.Context, id SubscriptionId, predicate PrivateCloudPredicate) (resp ListInSubscriptionCompleteResult, err error) {
	items := make([]PrivateCloud, 0)

	page, err := c.ListInSubscription(ctx, id)
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

	out := ListInSubscriptionCompleteResult{
		Items: items,
	}
	return out, nil
}

// preparerForListInSubscription prepares the ListInSubscription request.
func (c PrivateCloudsClient) preparerForListInSubscription(ctx context.Context, id SubscriptionId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/providers/Microsoft.AVS/privateClouds", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// preparerForListInSubscriptionWithNextLink prepares the ListInSubscription request with the given nextLink token.
func (c PrivateCloudsClient) preparerForListInSubscriptionWithNextLink(ctx context.Context, nextLink string) (*http.Request, error) {
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

// responderForListInSubscription handles the response to the ListInSubscription request. The method always
// closes the http.Response Body.
func (c PrivateCloudsClient) responderForListInSubscription(resp *http.Response) (result ListInSubscriptionResponse, err error) {
	type page struct {
		Values   []PrivateCloud `json:"value"`
		NextLink *string        `json:"nextLink"`
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
		result.nextPageFunc = func(ctx context.Context, nextLink string) (result ListInSubscriptionResponse, err error) {
			req, err := c.preparerForListInSubscriptionWithNextLink(ctx, nextLink)
			if err != nil {
				err = autorest.NewErrorWithError(err, "privateclouds.PrivateCloudsClient", "ListInSubscription", nil, "Failure preparing request")
				return
			}

			result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
			if err != nil {
				err = autorest.NewErrorWithError(err, "privateclouds.PrivateCloudsClient", "ListInSubscription", result.HttpResponse, "Failure sending request")
				return
			}

			result, err = c.responderForListInSubscription(result.HttpResponse)
			if err != nil {
				err = autorest.NewErrorWithError(err, "privateclouds.PrivateCloudsClient", "ListInSubscription", result.HttpResponse, "Failure responding to request")
				return
			}

			return
		}
	}
	return
}
