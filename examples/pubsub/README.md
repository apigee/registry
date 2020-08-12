# pubsub

This directory contains examples that use the Google PubSub API to receive
notifications from the `registry-server`. The `subscribe` directory contains an
event listener and the `publish` directory contains a tool that publishes
messages on the same topic (i.e. channel) used by `registry-server`.
