#!/bin/bash

# Script to add dependencies to database before running migration tests
# For example extensions, users, databases etc.

psql -v ON_ERROR_STOP=1 <<-EOSQL

EOSQL
