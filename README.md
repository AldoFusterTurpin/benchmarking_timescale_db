## Execute the Go program and TimescaleDB using docker-compose
```sh
docker compose up --build 
```
When executing the above command, if you find a log message like "PostgreSQL Database directory appears to contain a database; Skipping initialization", that means that the init script was skipped because there is already a DB persisted. In order to the init script to be aplied, you first need to remove the data appending the "--volumes" when shutting down the services:

```sh
docker compose down --volumes
```

## (Optional) Setup containerized DB instance to play with psql
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