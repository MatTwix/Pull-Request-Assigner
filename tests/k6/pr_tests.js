import http from 'k6/http';
import { check } from 'k6';
import { CONFIG } from './config.js';

export const options = {
  vus: 5,
  duration: "15s"
};

export default function () {
  let now = Date.now()

  for (let i = 1; i <= 4; i++) {
    const res = http.post(`${CONFIG.BASE_URL}/users/setIsActive`,
      JSON.stringify({
        user_id: `test_pr_user${i}${__VU}_${now}`,
        is_active: true
      }),
      {
        headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
      }
    );
  }

  const payload = JSON.stringify({
    team_name: `test_team${__VU}_${now}`,
    members: [
      { user_id: `test_pr_user1${__VU}_${now}`, username: "user", is_active: true },
      { user_id: `test_pr_user2${__VU}_${now}`, username: "user", is_active: true },
      { user_id: `test_pr_user3${__VU}_${now}`, username: "user", is_active: true },
      { user_id: `test_pr_user4${__VU}_${now}`, username: "user", is_active: true }
    ]
  });

  let res = http.post(`${CONFIG.BASE_URL}/team/add`, payload);

  const createRes = http.post(
    `${CONFIG.BASE_URL}/pullRequest/create`,
    JSON.stringify({
      pull_request_id: `test_pr${__VU}_${now}`,
      pull_request_name: "test_pr",
      author_id: `test_pr_user1${__VU}_${now}`
    }),
    { headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY } }
  );

  check(createRes, {
    "CREATE status 201": r => r.status === 201 || r.status === 400,
  });
  if (createRes.status != 201) console.log(createRes.status)

  if (createRes.json().pr.assigned_reviewers != null) {
    let reviewerId = createRes.json().pr.assigned_reviewers[0];

    const reassignRes = http.post(
      `${CONFIG.BASE_URL}/pullRequest/reassign`,
      JSON.stringify({
        pull_request_id: `test_pr${__VU}_${now}`,
        old_user_id: reviewerId
      }),
      { headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY } }
    );

    check(reassignRes, {
      "REASSIGN status 200": r => r.status === 200,
    });
    if (reassignRes.status != 200) console.log(reassignRes.status)
  }


  const mergeRes = http.post(`${CONFIG.BASE_URL}/pullRequest/merge`,
    JSON.stringify({ pull_request_id: `test_pr${__VU}_${now}` }),
    {
      headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    }
  );

  check(mergeRes, {
    "MERGE status 200": r => r.status === 200 || r.status === 400
  });
  if (mergeRes.status != 200) console.log(mergeRes.status)
}