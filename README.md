# How to reproduce 2 types of schema reloading related errors

1. To build the docker image of vitess based on Vitess version GA 8.0.0 that has 1 sharded keyspace called `test_sharded_keyspace` and 1 unsharded keyspace called `test_unsharded_keyspace`, run `docker build --no-cache -t vitess/test-database-2 .`
2. To run the docker container, run `docker run -p 15991:15991 -p 15306:15306 -p 15999:15999 vitess/test-database-2`
3. To create aliases, run `alias mysql='command mysql -h 127.0.0.1 -P 15306'; alias vtctlclient='command vtctlclient -server localhost:15999 -log_dir /vt/vtdataroot/tmp -alsologtostderr'`
4. To count the number of rows, run `mysql`, then `use test_sharded_keyspace;`, then `select count(1) from customer;`
5. To populate the table `test_sharded_keyspace.customer`, run `go run populate_vitess.go`, keep it running, while you follow the next steps
6. Wait until you have 50k~100k rows, then run `vtctlclient ApplySchema -sql "ALTER WITH 'gh-ost' TABLE customer add column x1 int;" test_sharded_keyspace`
7. Run `vtctlclient OnlineDDL test_sharded_keyspace show recent` repeatedly to check the `migration_status`
8. When `migration_status` is `running`, run `go run vstream_client.go MASTER` and `go run vstream_client.go REPLICA`, the 2 vstream clients failed with the corresponding following errors
```
Subscribe to tablet type: MASTER
..2020-12-21 18:59:34.634694 +0200 EET m=+4.299081337:: remote error: Code: UNKNOWN
rpc error: code = Unknown desc = target: test_sharded_keyspace.-80.master, used tablet: zone1-200 (c55f889869fb): vttablet: rpc error: code = Unknown desc = stream (at source tablet) error @ e0b72f55-43aa-11eb-a9f6-0242ac110002:1-590: unknown table _a519a318_43ad_11eb_84a6_0242ac110002_20201221165921_ghc in schema
```

```
Subscribe to tablet type: REPLICA
..2020-12-21 18:59:34.679474 +0200 EET m=+4.507299459:: remote error: Code: UNKNOWN
rpc error: code = Unknown desc = target: test_sharded_keyspace.80-.replica, used tablet: zone1-301 (c55f889869fb): vttablet: rpc error: code = Unknown desc = stream (at source tablet) error @ e89f88d1-43aa-11eb-9b0c-0242ac110002:1-583: unknown table _a519a318_43ad_11eb_84a6_0242ac110002_20201221165921_ghc in schema
```
9. A **FEW SECONDS** AFTER `migration_status` is `complete`, run `go run vstream_client.go MASTER` and `go run vstream_client.go REPLICA`, the 2 vstream clients failed with the corresponding following different errors
```
Subscribe to tablet type: MASTER
..2020-12-21 18:59:53.057705 +0200 EET m=+4.760761986:: remote error: Code: UNKNOWN
rpc error: code = Unknown desc = target: test_sharded_keyspace.80-.master, used tablet: zone1-300 (c55f889869fb): vttablet: rpc error: code = Unknown desc = stream (at source tablet) error @ e89f88d1-43aa-11eb-9b0c-0242ac110002:1-680: cannot determine table columns for customer: event has [8 15 3], schema as [name:"customer_id" type:INT64 table:"customer" org_table:"customer" database:"vt_test_sharded_keyspace" org_name:"customer_id" column_length:20 charset:63 flags:53251  name:"email" type:VARBINARY table:"customer" org_table:"customer" database:"vt_test_sharded_keyspace" org_name:"email" column_length:128 charset:63 flags:128 ]
```

```
Subscribe to tablet type: REPLICA
..2020-12-21 18:59:54.125235 +0200 EET m=+4.279017565:: remote error: Code: UNKNOWN
rpc error: code = Unknown desc = target: test_sharded_keyspace.80-.replica, used tablet: zone1-301 (c55f889869fb): vttablet: rpc error: code = Unknown desc = stream (at source tablet) error @ e89f88d1-43aa-11eb-9b0c-0242ac110002:1-680: cannot determine table columns for customer: event has [8 15 3], schema as [name:"customer_id" type:INT64 table:"customer" org_table:"customer" database:"vt_test_sharded_keyspace" org_name:"customer_id" column_length:20 charset:63 flags:53251  name:"email" type:VARBINARY table:"customer" org_table:"customer" database:"vt_test_sharded_keyspace" org_name:"email" column_length:128 charset:63 flags:128 ]
```