import http from 'k6/http';
import { check, sleep } from 'k6';
import { CONFIG } from './config.js';

export const options = {
  vus: 10,
  duration: '15s',
};

export default function () {
  let now = (Date.now() / 10) | 0

  const payload = JSON.stringify({
    team_name: `test_team${__VU}_${now}`,
    members: [
      { user_id: `test_user1${__VU}_${now}`, username: "user", is_active: true },
      { user_id: `test_user2${__VU}_${now}`, username: "user", is_active: true },
      { user_id: `test_user3${__VU}_${now}`, username: "user", is_active: true },
      { user_id: `test_user4${__VU}_${now}`, username: "user", is_active: true },
      { user_id: `test_user5${__VU}_${now}`, username: "user", is_active: true },
    ]
  });

  let res = http.post(`${CONFIG.BASE_URL}/team/add`, payload);

  check(res, { "ADD team status 201": (r) => r.status === 201 });
  if (res.status != 201) console.log(res.status)

  res = http.get(`${CONFIG.BASE_URL}/team/get?team_name=test_team${__VU}_${now}`, {
    headers: { "X-Api-Key": CONFIG.ADMIN_API_KEY }
  });

  check(res, { "GET team status 200": r => r.status === 200 });
  if (res.status != 200) console.log(res.status)

  let deactivateRes = http.post(`${CONFIG.BASE_URL}/team/deactivate`,
    JSON.stringify({ team_name: `test_team${__VU}_${now}` }),
    {
      headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    }
  );

  check(deactivateRes, { "DEACTIVATE status 200": r => r.status === 200 });
  if (deactivateRes.status != 200) console.log(deactivateRes.status)
}