package nodetype

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type ListByManagedClustersResponse struct {
	HttpResponse *http.Response
	Model        *[]NodeType

	nextLink     *string
	nextPageFunc func(ctx context.Context, nextLink string) (ListByManagedClustersResponse, error)
}

type ListByManagedClustersCompleteResult struct {
	Items []NodeType
}

func (r ListByManagedClustersResponse) HasMore() bool {
	return r.nextLink != nil
}

func (r ListByManagedClustersResponse) LoadMore(ctx context.Context) (resp ListByManagedClustersResponse, err error) {
	if !r.HasMore() {
		err = fmt.Errorf("no more pages returned")
		return
	}
	return r.nextPageFunc(ctx, *r.nextLink)
}

// ListByManagedClusters ...
func (c NodeTypeClient) ListByManagedClusters(ctx context.Context, id ManagedClusterId) (resp ListByManagedClustersResponse, err error) {
	req, err := c.preparerForListByManagedClusters(ctx, id)
	if err != nil {
		err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "ListByManagedClusters", nil, "Failure preparing request")
		return
	}

	resp.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
	if err != nil {
		err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "ListByManagedClusters", resp.HttpResponse, "Failure sending request")
		return
	}

	resp, err = c.responderForListByManagedClusters(resp.HttpResponse)
	if err != nil {
		err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "ListByManagedClusters", resp.HttpResponse, "Failure responding to request")
		return
	}
	return
}

// ListByManagedClustersComplete retrieves all of the results into a single object
func (c NodeTypeClient) ListByManagedClustersComplete(ctx context.Context, id ManagedClusterId) (ListByManagedClustersCompleteResult, error) {
	return c.ListByManagedClustersCompleteMatchingPredicate(ctx, id, NodeTypePredicate{})
}

// ListByManagedClustersCompleteMatchingPredicate retrieves all of the results and then applied the predicate
func (c NodeTypeClient) ListByManagedClustersCompleteMatchingPredicate(ctx context.Context, id ManagedClusterId, predicate NodeTypePredicate) (resp ListByManagedClustersCompleteResult, err error) {
	items := make([]NodeType, 0)

	page, err := c.ListByManagedClusters(ctx, id)
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

	out := ListByManagedClustersCompleteResult{
		Items: items,
	}
	return out, nil
}

// preparerForListByManagedClusters prepares the ListByManagedClusters request.
func (c NodeTypeClient) preparerForListByManagedClusters(ctx context.Context, id ManagedClusterId) (*http.Request, error) {
	queryParameters := map[string]interface{}{
		"api-version": defaultApiVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsGet(),
		autorest.WithBaseURL(c.baseUri),
		autorest.WithPath(fmt.Sprintf("%s/nodeTypes", id.ID())),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// preparerForListByManagedClustersWithNextLink prepares the ListByManagedClusters request with the given nextLink token.
func (c NodeTypeClient) preparerForListByManagedClustersWithNextLink(ctx context.Context, nextLink string) (*http.Request, error) {
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

// responderForListByManagedClusters handles the response to the ListByManagedClusters request. The method always
// closes the http.Response Body.
func (c NodeTypeClient) responderForListByManagedClusters(resp *http.Response) (result ListByManagedClustersResponse, err error) {
	type page struct {
		Values   []NodeType `json:"value"`
		NextLink *string    `json:"nextLink"`
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
		result.nextPageFunc = func(ctx context.Context, nextLink string) (result ListByManagedClustersResponse, err error) {
			req, err := c.preparerForListByManagedClustersWithNextLink(ctx, nextLink)
			if err != nil {
				err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "ListByManagedClusters", nil, "Failure preparing request")
				return
			}

			result.HttpResponse, err = c.Client.Send(req, azure.DoRetryWithRegistration(c.Client))
			if err != nil {
				err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "ListByManagedClusters", result.HttpResponse, "Failure sending request")
				return
			}

			result, err = c.responderForListByManagedClusters(result.HttpResponse)
			if err != nil {
				err = autorest.NewErrorWithError(err, "nodetype.NodeTypeClient", "ListByManagedClusters", result.HttpResponse, "Failure responding to request")
				return
			}

			return
		}
	}
	return
}
