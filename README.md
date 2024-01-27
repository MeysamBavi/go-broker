# Message broker in Go

## Overview
This project implements a simple message broker in **Go**, with three **gRPC** endpoints: *publish*, *fetch*, and *subscribe*. It features a clean and efficient codebase, supports flexible storage, and is containerized and deployed using **Docker** and **Kubernetes**.  
This was part of *Bale* messenger's bootcamp.

## Architecture
<div align="center">

![Software Architecture](assets/arch.svg)

<p>Each published message is stored and broadcast to all subscribers</p>

</div>

## Key Features

- **Storage Flexibility**: Utilizes three storage approaches; in-memory, **PostgreSQL**, and **Cassandra**

- **Containerization and Deployment**:
  - Leverages Docker for containerization
  - Incorporates Kubernetes for deployment, including resources and *Bash* scripts for seamless application setup and teardown

- **Clean and Efficient Codebase**:
  - Written in Go, emphasizing clean code and high efficiency
  - Utilizes interfaces for flexibility and modularity
  - Organizes code into different packages and layers for improved modularity

- **Monitoring and Metrics**:
  - Employs **Prometheus** for comprehensive metric solutions
  - Integrates **Grafana** for intuitive visualization of performance metrics

- **Tracing**:
  - Uses **Jaeger** for tracing.
  - Utilizes the **OpenTracing** library in Go for creating spans and collecting trace data

- **Rate Limiting**: Uses **Envoy proxy** for traffic management and rate limiting

- **Load Testing**:
  - Employs **k6** for load testing, measuring performance across different storage technologies
  - Includes a **Go gRPC client** for load testing, as a superior alternative to k6 scripts

- **Optimization through Batch Creation**:
  - Leverages *'batch creation'* method to optimize the *publish* procedure during high insertion loads
  - Modular batch logic applicable across various storage technologies as a reusable dependency

## Data Model
- Unique ID per subject (topic) for stored messages
- Time-to-live for messages, ensuring expiration after a specified duration

## How to run
### Docker
Select one of different docker compose files, depending on your intended configuration; For example `docker-compose-postgres-loadtest.yaml`. Then run
```shell
docker compose -f docker-compose-postgres-loadtest.yaml up
```
### Kubernetes
```shell
./k8s/up.sh
```
Grafana can be accessed on `host:3000`.
