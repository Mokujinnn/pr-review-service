#!/bin/sh

set -e

host="$1"
shift
port="$1"
shift
cmd="$@"

until nc -z "$host" "$port"; do
  echo "Waiting for database $host:$port..."
  sleep 1
done

echo "Database is ready! Starting application..."
exec $cmd