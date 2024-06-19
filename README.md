### Problem statement
To implement a command line tool that can be used to benchmark SELECT query performance across multiple workers/clients against a TimescaleDB instance. The tool should take as its input a CSV-formatted file and a flag to specify the number of concurrent workers. The tool should gracefully handle files that are larger than the one given, and should not wait to start processing queries until all input is consumed. After processing all the queries specified by the parameters in the CSV file, the tool should output a summary with the following stats: 
- number of queries processed
- total processing time across all queries
- the minimum query time (for a single query)
- the median query time
- the average query time
- and the maximum query time

### Strategy chosen
I have decided to use a worker pool with one go routine acting as a multiplexer/coordinator and n workers that will process the requests. I have used a chanel of channels where each value "q" of this chanel is itself a chanel of Request. 

Each of the workers will be waiting for requests in one specific chanel "q" that will contain all the requests for the same hostname, and once a request is ready for that worker, it will process it and send the result to the corresponding result chanel (result chanel is the same for all workers but could be different). 

I have used the modulo operator (%) to know which workerId corresponds to which hostname, as we must process the request for the same hostname by the same worker (constraint imposed by the problem statement). Due to the previous constraint, the number of workers that the program will span will 
be minimum(nWorkers, nHostnames) where nWorkers is specified in an env variable ("WORKERS", or default value is used) and nHostnames is determined by the value in query_params.csv, as it make no sense to span more workers than number of hostnames we have because we would have workers doing nothing in such case.
Ex:

hostnames:

hostname_1 hostname_2 hostname_3 hostname_4 hostname_5

if nWorkers was 10:

all requests of hostname_1 -> goes to worker 0

all requests of hostname_2 -> goes to worker 1

all requests of hostname_3 -> goes to worker 2 

all requests of hostname_4 -> goes to worker 3

all requests of hostname_5 -> goes to worker 4

and we would have workers 5 to 10 doing nothing as there is no hostname for them, because same hostname must goe to the same worker.

That is why the number of workers I span is min(nWorkers, nHostnames).

If we have more hosts than workers, the hosts are distributed to the workers:
Ex:
hostnames:

hostname_1 hostname_2 hostname_3 hostname_4 hostname_5

if nWorkers was 2

all requests of hostname_1 -> goes to worker 0

all requests of hostname_2 -> goes to worker 1

all requests of hostname_3 -> goes to worker 0

all requests of hostname_4 -> goes to worker 1

all requests of hostname_5 -> goes to worker 0


I have used a map to know at which chanel each request should go, to ensure each requests will be processed by the corresponding worker (in conjunction with the above strategy explained).

Each worker will start to process requests as soon as it reads any from the corresponding chanel, and will send the results as soon as it has process the request. Once we have sending all the requests, we can close those channels, and once all the requests have been processed, we can close the result chanel.

The multiplexer runs in a different goroutine and all the workers are waiting for more work (reading from its corresponding chanel) in a different goroutine. The retrieval of the results is also run in a different goroutine.

### Execute the Go program and TimescaleDB using docker-compose
```sh
make run
```
When executing the above command, if you find a log message like "PostgreSQL Database directory appears to contain a database; Skipping initialization", that means that the init script was skipped because there is already a DB persisted. In order to the init script to be aplied, you first need to remove the data appending the "--volumes" when shutting down the services, already done by the following make directive:

```sh
make clean
```

If you want to specify the number of workers, change the value of compose.yaml > app > environment > WORKERS to the specific value, like:
```sh
environment:
  - WORKERS=10
```

If no value is specified or the number is not a valid integer, the program will use a value by default (5).

### Run the unit tests
```sh
make test
```

### (Optional) Setup containerized DB instance to play with psql
Pull DB image:
```sh
docker pull timescale/timescaledb-ha:pg16
```

Run the DB container:
```sh
docker run -d --name timescaledb -p 5432:5432 -e POSTGRES_PASSWORD=password timescale/timescaledb-ha:pg16
```

Connect to a database on the PostgreSQL instance:

```sh
psql -d "postgres://postgres:password@localhost:5432/postgres"
```

the format is

```sh
psql -d "postgres://<username>:<password>@<host>:<port>/<database-name>"
```

and the default user and database are both "postgres", so doing
```sh
psql -d "postgres://postgres:password@localhost:5432"
```
will connect using user "postgres" and password "password" to the default database "postgres"

Check that TimescaleDB is installed:
```sh
\dx
```

And we can exit:
```sh
\q
```
Now we execute the following query to set a new database called "homework" and create a hypertable called "cpu_usage":

```sh
psql -d "postgres://postgres:password@localhost:5432" < cpu_usage.sql
```

We need to populate the table with data (note the **/homework** at the end of the connection string):
```sh
psql -d "postgres://postgres:password@localhost:5432/homework" -c "\COPY cpu_usage FROM cpu_usage.csv CSV HEADER"
```

Note that from now onwards we will be working with the "homework" DB and "cpu_usage" table, so our connection string to PosgtreSQL will be:
```sh
psql -d "postgres://postgres:password@localhost:5432/homework"
```
Note: in the Go program, the connection string is different as we need to reference the host by the name of the service specified in the compose.yaml file, and NOT localhost.

Now we can use the previous command to play with the DB.