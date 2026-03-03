import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10000, // Number of virtual users      
  duration: '30s',   
};

export default function () {
  const res = http.post('http://localhost:3123/ingest', JSON.stringify({
    timestamp: new Date().toISOString(),
    level: 'info',
    message: 'This is a test log message',
    metadata: {
      userId: Math.floor(Math.random() * 1000),
      sessionId: Math.random().toString(36).substring(2, 15),
    },
  }), {
    headers: { 'Content-Type': 'application/json' },
  })    ;

    check(res, {'status is 202': (r) => r.status === 202});

  sleep(1);
}