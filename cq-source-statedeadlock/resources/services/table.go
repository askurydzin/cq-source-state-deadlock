package services

import (
	"context"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/cloudquery/plugin-sdk/v4/schema"
)

func SampleTable() *schema.Table {
	return &schema.Table{
		Name:     "statedeadlock_sample_table",
		Resolver: fetchSampleTable,
		Columns: []schema.Column{
			{
				Name: "column",
				Type: arrow.BinaryTypes.String,
			},
		},
	}
}

func fetchSampleTable(ctx context.Context, meta schema.ClientMeta, parent *schema.Resource, res chan<- any) error {
	return nil
}
