// k6: 20k logs/sec burst for 30s
// TODO TASK-070: implement k6 burst test
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 200,
  duration: '30s',
  thresholds: { http_req_failed: ['rate<0.05'] },
};

export default function () {
  http.post(`${__ENV.INGEST_URL}/v1/ingest/logs`, JSON.stringify({
    service: 'checkout-svc',
    entries: [{ level: 'ERROR', message: 'burst test', timestamp: new Date().toISOString() }],
  }), { headers: { 'Content-Type': 'application/json', 'X-API-Key': __ENV.API_KEY } });
  sleep(0.01);
}
