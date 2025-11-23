import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 20,
  duration: '30s',
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.001'],    
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://app:8080';

export default function () {
  let res = http.get(`${BASE_URL}/health`);
  check(res, {
    'health is 200': (r) => r.status === 200,
  });

  res = http.get(`${BASE_URL}/users/getReview?user_id=u2`);
  check(res, {
    'getReview is 200': (r) => r.status === 200,
  });

  sleep(0.1);
}
