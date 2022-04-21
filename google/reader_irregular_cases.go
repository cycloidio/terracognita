package google

import (
	"context"

	"strings"

	"github.com/pkg/errors"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// ListSQLDatabases returns a list of Databases within a project and a instances
func (r *GCPReader) ListSQLDatabases(ctx context.Context, filter string, instances []string) (map[string][]sqladmin.Database, error) {

	service := sqladmin.NewDatabasesService(r.sqladmin)
	list := make(map[string][]sqladmin.Database)
	for _, elem := range instances {
		resources := make([]sqladmin.Database, 0)

		elemList, err := service.List(r.project, elem).Context(ctx).Do()
		if err != nil {
			//if 404 remove because resource instance is stop - otherwise fails
			if strings.Contains(err.Error(), "Invalid request since instance is not running.") {
				continue
			}
			//if resource instance is on-premise - otherwise fails
			if strings.Contains(err.Error(), "400") {
				continue
			}
			return nil, errors.Wrap(err, "unable to list sqladmin Database from google APIs")
		}

		for _, res := range elemList.Items {
			resources = append(resources, *res)
		}

		list[elem] = resources
	}
	return list, nil

}
