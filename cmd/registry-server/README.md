# registry-server

This directory contains the main entry point for running the Registry API
server. To support running in certain hosted environments, it uses the `PORT`
environment variable to determine on which port to run.

In hosted Google environments, it receives all other configuration from
automatically-provided environment variables. In other enviroments (including
when run locally), `registry-server` requires database configuration as
described in the top-level [README](/README.md) of this repo.
