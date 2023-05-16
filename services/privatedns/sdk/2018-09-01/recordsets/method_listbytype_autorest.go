package recordsets

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListByTypeOperationResponse struct {
	HttpResponse *http.Response
	Model        *[]RecordSet

	nextLink     *string
	nextPageFunc func(ctx context.Context, nextLink string) (ListByTypeOperationResponse, error)
}

type ListByTypeCompleteResult struct {
	Items []RecordSet
}

func (r ListByTypeOperationResponse) HasMore() bool {
	return r.nextLink != nil
}

func (r ListByTypeOperationResponse) LoadMore(ctx context.Context) (resp ListByTypeOperationResponse, err error) {
	if !r.HasMore() {
		err = fmt.Errorf("no more pages returned")
		return
	}
	return r.nextPageFunc(ctx, *r.nextLink)
}

type ListByTypeOperationOptions struct {
	Recordsetnamesuffix *string
	Top                 *int64
}

func DefaultListByTypeOperationOptions() ListByTypeOperationOptions {
	return ListByTypeOperationOptions{}
}

func (o ListByTypeOperationOptions) toHeaders() map[string]interface{} {
	out := make(map[string]interface{})

	return out
}

func (o ListByTypeOperationOptions) toQueryString() map[string]interface{} {
	out := make(map[string]interface{})

	if o.Recordsetnamesuffix != nil {
		out["$recordsetnamesuffix"] = *o.Recordsetnamesuffix
	}

	if o.Top != nil {
		out["$top"] = *o.Top
	}

	return out
}

// ListByType ...
func (c RecordSetsClient) ListByType(ctx context.Context, id PrivateZoneId, options ListByTypeOperationOptions) (resp ListByTypeOperationResponse, err error) {
	req, err := c.preparerForListByType(ctx, id, options)
	if err != nil {
		err = autorest.NewErrorWithError(err, "recordsets.RecordSetsClient", "ListByType", nil, "Failure preparing request")
		return
	}

	resp.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "recordsets.RecordSetsClient", "ListByType", resp.HttpResponse, "Failure sending request")
		return
	}

	resp, err = c.responderForListByType(resp.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "recordsets.RecordSetsClient", "ListByType", resp.HttpResponse, "Failure responding to request")
		return
	}
	return
}

// ListByTypeComplete retrieves all of the results into a single object
func (c RecordSetsClient) ListByTypeComplete(ctx context.Context, id PrivateZoneId, options ListByTypeOperationOptions) (ListByTypeCompleteResult, error) {
	return c.ListByTypeCompleteMatchingPredicate(ctx, id, options, RecordSetOperationPredicate{})
}

// ListByTypeCompleteMatchingPredicate retrieves all of the results and then applied the predicate
func (c RecordSetsClient) ListByTypeCompleteMatchingPredicate(ctx context.Context, id PrivateZoneId, options ListByTypeOperationOptions, predicate RecordSetOperationPredicate) (resp ListByTypeCompleteResult, err error) {
	items := make([]RecordSet, 0)

	page, err := c.ListByType(ctx, id, options)
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

	out := ListByTypeCompleteResult{
		Items: items,
	}
	return out, nil
}

// preparerForListByType prepares the ListByType request.
func (c RecordSetsClient) preparerForListByType(ctx context.Context, id PrivateZoneId, options ListByTypeOperationOptions) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	for k, v := range options.toQueryString() {
		queryParameters[k] = autorest.Encode("query", v)
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithHeaders(options.toHeaders()),
		autorest.WithPath(id.ID()),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// preparerForListByTypeWithNextLink prepares the ListByType request with the given nextLink token.
func (c RecordSetsClient) preparerForListByTypeWithNextLink(ctx context.Context, nextLink string) (*http.Request, error) {
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

// responderForListByType handles the response to the ListByType request. The method always
// closes the http.Response Body.
func (c RecordSetsClient) responderForListByType(resp *http.Response) (result ListByTypeOperationResponse, err error) {
	type page struct {
		Values   []RecordSet `json:"value"`
		NextLink *string     `json:"nextLink"`
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
		result.nextPageFunc = func(ctx context.Context, nextLink string) (result ListByTypeOperationResponse, err error) {
			req, err := c.preparerForListByTypeWithNextLink(ctx, nextLink)
			if err != nil {
				err = autorest.NewErrorWithError(err, "recordsets.RecordSetsClient", "ListByType", nil, "Failure preparing request")
				return
			}

			result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
			if err != nil {
				err = autorest.NewErrorWithError(err, "recordsets.RecordSetsClient", "ListByType", result.HttpResponse, "Failure sending request")
				return
			}

			result, err = c.responderForListByType(result.HttpResponse)
			if err != nil {
				err = autorest.NewErrorWithError(err, "recordsets.RecordSetsClient", "ListByType", result.HttpResponse, "Failure responding to request")
				return
			}

			return
		}
	}
	return
}
