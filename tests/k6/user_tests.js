import http from 'k6/http';
import { check } from 'k6';
import { CONFIG } from './config.js';

export const options = {
  vus: 10,
  duration: "15s"
};

export default function () {
  const today = new Date();
  const dateStr = today.toISOString().split('T')[0];

  const res = http.post(`${CONFIG.BASE_URL}/users/setIsActive`,
    JSON.stringify({
      user_id: "smoke_user1",
      is_active: Math.random() < 0.5
    }),
    {
      headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    }
  );

  check(res, { "SETISACTIVE status 200": r => r.status === 200 });
  if (res.status != 200) console.log(res.status)

  let getReviewRes = http.get(`${CONFIG.BASE_URL}/users/getReview?user_id=test_user${Math.floor(Math.random() * 6) + 1}`, {
    headers: { "X-Api-Key": CONFIG.ADMIN_API_KEY }
  });

  check(getReviewRes, { "GETREVIEW status 200": r => r.status === 200 });
  if (getReviewRes.status != 200) console.log(getReviewRes.status)
}