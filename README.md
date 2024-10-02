# State Table Dead-Lock Sample

```
# build cq plugin
make build-cq

# up postgresql database
make up

# run cloudquery sync
go run main.go
```


# Explanation

The plugin contains ridiculous code that causes the dead-lock situation to happen:

```
    // update 10000 keys in random order.
	for _, i := range rand.Perm(10000) {
		stateClient.SetKey(ctx, fmt.Sprintf("key%d", i), fmt.Sprintf("value%d-%d", i, rand.Int63()))
	}

    // run sync no-op.
	if err := c.scheduler.Sync(ctx, c.syncClient, tt, res, scheduler.WithSyncDeterministicCQID(options.DeterministicCQID)); err != nil {
		return err
	}

    // flush keys (dead-lock).
	return stateClient.Flush(ctx)
```

# Sample Output

```
$ go run main.go 
exit status 1
2024-10-02T12:13:36Z ERR failed to write batch error="failed to execute batch with pgerror: severity: ERROR, code: 40P01, message: deadlock detected, detail :Process 99 waits for ShareLock on transaction 499; blocked by process 98.\nProcess 98 waits for ShareLock on transaction 500; blocked by process 99., hint: See server log for query details., position: 0, internal_position: 0, internal_query: , where: while inserting index tuple (2,112) in relation \"deadlock_state\", schema_name: , table_name: , column_name: , data_type_name: , constraint_name: , file: deadlock.c, line: 1150, routine: DeadLockReport: ERROR: deadlock detected (SQLSTATE 40P01)" duration=1045.34281 len=10000 module=pg-dest
2024-10-02T12:13:36Z ERR finished call grpc.code=Internal grpc.component=server grpc.error="rpc error: code = Internal desc = write failed: failed to execute batch with pgerror: severity: ERROR, code: 40P01, message: deadlock detected, detail :Process 99 waits for ShareLock on transaction 499; blocked by process 98.\nProcess 98 waits for ShareLock on transaction 500; blocked by process 99., hint: See server log for query details., position: 0, internal_position: 0, internal_query: , where: while inserting index tuple (2,112) in relation \"deadlock_state\", schema_name: , table_name: , column_name: , data_type_name: , constraint_name: , file: deadlock.c, line: 1150, routine: DeadLockReport: ERROR: deadlock detected (SQLSTATE 40P01)" grpc.method=Write grpc.method_type=client_stream grpc.service=cloudquery.plugin.v3.Plugin grpc.start_time=2024-10-02T14:13:35+02:00 grpc.time_ms=1048.572 module=cli peer.address=@ protocol=grpc
2024-10-02T12:13:36Z ERR finished call grpc.code=Internal grpc.component=server grpc.error="failed to sync records: failed to sync unmanaged client: rpc error: code = Internal desc = write failed: failed to execute batch with pgerror: severity: ERROR, code: 40P01, message: deadlock detected, detail :Process 99 waits for ShareLock on transaction 499; blocked by process 98.\nProcess 98 waits for ShareLock on transaction 500; blocked by process 99., hint: See server log for query details., position: 0, internal_position: 0, internal_query: , where: while inserting index tuple (2,112) in relation \"deadlock_state\", schema_name: , table_name: , column_name: , data_type_name: , constraint_name: , file: deadlock.c, line: 1150, routine: DeadLockReport: ERROR: deadlock detected (SQLSTATE 40P01)" grpc.method=Sync grpc.method_type=server_stream grpc.service=cloudquery.plugin.v3.Plugin grpc.start_time=2024-10-02T14:13:34+02:00 grpc.time_ms=2158.577 module=cli peer.address=@ protocol=grpc
Error: failed to sync v3 source state-deadlock: unexpected error from sync client receive: rpc error: code = Internal desc = failed to sync records: failed to sync unmanaged client: rpc error: code = Internal desc = write failed: failed to execute batch with pgerror: severity: ERROR, code: 40P01, message: deadlock detected, detail :Process 99 waits for ShareLock on transaction 499; blocked by process 98.
Process 98 waits for ShareLock on transaction 500; blocked by process 99., hint: See server log for query details., position: 0, internal_position: 0, internal_query: , where: while inserting index tuple (2,112) in relation "deadlock_state", schema_name: , table_name: , column_name: , data_type_name: , constraint_name: , file: deadlock.c, line: 1150, routine: DeadLockReport: ERROR: deadlock detected (SQLSTATE 40P01)
2024-10-02T12:13:36Z ERR exiting with error error="failed to sync v3 source state-deadlock: unexpected error from sync client receive: rpc error: code = Internal desc = failed to sync records: failed to sync unmanaged client: rpc error: code = Internal desc = write failed: failed to execute batch with pgerror: severity: ERROR, code: 40P01, message: deadlock detected, detail :Process 99 waits for ShareLock on transaction 499; blocked by process 98.\nProcess 98 waits for ShareLock on transaction 500; blocked by process 99., hint: See server log for query details., position: 0, internal_position: 0, internal_query: , where: while inserting index tuple (2,112) in relation \"deadlock_state\", schema_name: , table_name: , column_name: , data_type_name: , constraint_name: , file: deadlock.c, line: 1150, routine: DeadLockReport: ERROR: deadlock detected (SQLSTATE 40P01)" module=cli
```
