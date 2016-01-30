#!/usr/bin/env bash

go build -o speed
./speed -file s6.png -cpuprofile=p.proff
go tool pprof speed p.proff
rm speed
rm p.proff
