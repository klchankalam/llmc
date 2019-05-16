@echo off
docker build -t llmc .
docker run -p 8080:8080 -it llmc