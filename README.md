# Flowbeat

Flowbeat is the [Beat](https://www.elastic.co/products/beats) used for
collecting sflow data.

# Current Status
Flowbeat at its current state (2016-02-17) works for router generated Sflow data and imports the following extended Flow Samples:
 - Basic Link Layer Headers and Protocols (IP, TCP, UDP)
 - ExtendedGateway Flows
 - ExtendedSwitch Flows
 - ExtendedRouter Flows

The used sflow library also parses several Host s-flow samples but this is untested.

Generally if flowbeat does not show the samples you want, the sflow library is probably lacking parser support for them.

## Documentation

Set the correct Listen port in flowbeat.yml and start sending it sflow packets.

TODO

## Exported fields

Currently flowbeat exports the raw parsed sflow data. Exactly as it is received and parsed from the wire

