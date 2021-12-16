#!/bin/bash
exe="./exedir/main"

go build -o $exe &&
$exe $@