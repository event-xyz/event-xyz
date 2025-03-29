# Eventloop Monitoring

This directory contains the configuration files for monitoring eloop with Prometheus and Grafana.

## Getting Started

1. Start the monitoring stack with Docker Compose:

```bash
docker-compose -f docker-compose.monitoring.yml up -d
```

2. Access the services:

- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (login: admin/admin)

3. The Grafana dashboard "Eventloop Dashboard" should be automatically provisioned.

## Metrics Available

The following metrics are available:

- `eventloop_up_since_seconds`: Application start time
- `eventloop_http_requests_total`: Total HTTP requests count by method, endpoint, and status
- `eventloop_http_request_duration_seconds`: HTTP request duration histogram
- `eventloop_participants_active`: Currently active participants gauge
- `eventloop_participants_checkin_total`: Total participant check-ins by team
- `eventloop_participants_checkout_total`: Total participant check-outs by team
- `eventloop_checkpoint_crossings_total`: Total checkpoint crossings by checkpoint
- `eventloop_qr_failures_total`: QR processing failures by reason
- `eventloop_auth_failures_total`: Authentication failures by role and reason
- `eventloop_db_errors_total`: Database errors by operation and error type

## Customizing Dashboards

You can customize the dashboards by:

1. Using the Grafana UI to modify the existing dashboard
2. Saving the dashboard JSON and replacing the existing file
