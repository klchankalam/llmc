#!/bin/bash
docker build -t llmc .
RESULT=$?
if [ $RESULT -eq 0 ]; then
  docker run -p 8080:8080 -it llmc
else
  echo failed
fi