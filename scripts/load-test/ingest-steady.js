// k6: constant 10k logs/sec for 5 min
// TODO TASK-070: implement k6 load test
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 50,
  duration: '5m',
  thresholds: { http_req_failed: ['rate<0.01'] },
};

export default function () {
  http.post(`${__ENV.INGEST_URL}/v1/ingest/logs`, JSON.stringify({
    service: 'checkout-svc',
    entries: [{ level: 'INFO', message: 'test', timestamp: new Date().toISOString() }],
  }), { headers: { 'Content-Type': 'application/json', 'X-API-Key': __ENV.API_KEY } });
  sleep(0.005);
}
