package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/askurydzin/cq-source-state-deadlock/cq-source-statedeadlock/client"
	"github.com/askurydzin/cq-source-state-deadlock/cq-source-statedeadlock/resources/services"
	"github.com/cloudquery/plugin-sdk/v4/message"
	"github.com/cloudquery/plugin-sdk/v4/plugin"
	"github.com/cloudquery/plugin-sdk/v4/scheduler"
	"github.com/cloudquery/plugin-sdk/v4/schema"
	"github.com/cloudquery/plugin-sdk/v4/state"
	"github.com/cloudquery/plugin-sdk/v4/transformers"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	logger     zerolog.Logger
	config     client.Spec
	tables     schema.Tables
	syncClient *client.Client
	scheduler  *scheduler.Scheduler

	plugin.UnimplementedDestination
}

const maxMsgSize = 100 * 1024 * 1024 // 100 MiB

func Configure(ctx context.Context, logger zerolog.Logger, spec []byte, opts plugin.NewClientOptions) (plugin.Client, error) {
	if opts.NoConnection {
		return &Client{
			logger: logger,
			tables: getTables(),
		}, nil
	}

	config := &client.Spec{}
	if err := json.Unmarshal(spec, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
	}

	syncClient, err := client.New(ctx, logger, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	return &Client{
		logger:     logger,
		config:     *config,
		tables:     getTables(),
		syncClient: &syncClient,
		scheduler:  scheduler.NewScheduler(scheduler.WithLogger(logger)),
	}, nil
}

func (c *Client) Sync(ctx context.Context, options plugin.SyncOptions, res chan<- message.SyncMessage) error {
	tt, err := c.tables.FilterDfs(options.Tables, options.SkipTables, options.SkipDependentTables)
	if err != nil {
		return err
	}

	var stateClient state.Client

	if options.BackendOptions == nil || options.BackendOptions.Connection == "" || options.BackendOptions.TableName == "" {
		c.logger.Info().Msg("No backend options provided, using no state backend")
		stateClient = &state.NoOpClient{}
	} else {
		conn, err := grpc.DialContext(ctx, options.BackendOptions.Connection,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(maxMsgSize),
				grpc.MaxCallSendMsgSize(maxMsgSize),
			),
		)
		if err != nil {
			return fmt.Errorf("failed to dial grpc source plugin at %s: %w", options.BackendOptions.Connection, err)
		}
		stateClient, err = state.NewClient(ctx, conn, options.BackendOptions.TableName)
		if err != nil {
			return fmt.Errorf("failed to create state client: %w", err)
		}
	}

	rand.Seed(time.Now().UnixNano())

	for _, i := range rand.Perm(10000) {
		stateClient.SetKey(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d-%d", i, rand.Int63()))
	}

	if err := c.scheduler.Sync(ctx, c.syncClient, tt, res, scheduler.WithSyncDeterministicCQID(options.DeterministicCQID)); err != nil {
		return err
	}
	return stateClient.Flush(ctx)
}

func (c *Client) Tables(_ context.Context, options plugin.TableOptions) (schema.Tables, error) {
	tt, err := c.tables.FilterDfs(options.Tables, options.SkipTables, options.SkipDependentTables)
	if err != nil {
		return nil, err
	}

	return tt, nil
}

func (*Client) Close(_ context.Context) error {
	// TODO: Add your client cleanup here
	return nil
}

func getTables() schema.Tables {
	tables := schema.Tables{
		services.SampleTable(),
	}
	if err := transformers.TransformTables(tables); err != nil {
		panic(err)
	}
	for _, t := range tables {
		schema.AddCqIDs(t)
	}
	return tables
}
