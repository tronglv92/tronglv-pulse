// k6: 100 concurrent WebSocket connections
// TODO TASK-070: implement k6 WebSocket test
import ws from 'k6/ws';
import { check, sleep } from 'k6';

export const options = {
  vus: 100,
  duration: '2m',
};

export default function () {
  const res = ws.connect(`${__ENV.WS_URL}/v1/stream`, {}, function (socket) {
    socket.on('open', () => socket.send(JSON.stringify({ type: 'subscribe', tenantId: 'demo' })));
    socket.on('message', (data) => check(data, { 'valid json': (d) => { try { JSON.parse(d); return true; } catch { return false; } } }));
    socket.setTimeout(() => socket.close(), 60000);
  });
  check(res, { 'connected': (r) => r && r.status === 101 });
}
