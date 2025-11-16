import http from 'k6/http';
import { check, sleep } from 'k6';
import { CONFIG } from './config.js';

export const options = {
    vus: 1,
    iterations: 1,
};

export default function () {
    // TEAM ADD
    let now = Date.now()

    let res = http.post(`${CONFIG.BASE_URL}/team/add`, JSON.stringify({
        team_name: `team_smoke${now}`,
        members: [
            { user_id: "smoke_user1", username: "user", is_active: true },
            { user_id: "smoke_user2", username: "user", is_active: true },
            { user_id: "smoke_user3", username: "user", is_active: true },
            { user_id: "smoke_user4", username: "user", is_active: true },
            { user_id: "smoke_user5", username: "user", is_active: true },
            { user_id: "smoke_user6", username: "user", is_active: true },
            { user_id: "smoke_user11", username: "user", is_active: true },
            { user_id: "smoke_user21", username: "user", is_active: true },
            { user_id: "smoke_user31", username: "user", is_active: true },
            { user_id: "smoke_user41", username: "user", is_active: true },
            { user_id: "smoke_user51", username: "user", is_active: true },
            { user_id: "smoke_user61", username: "user", is_active: true },
            { user_id: "smoke_user12", username: "user", is_active: true },
            { user_id: "smoke_user22", username: "user", is_active: true },
            { user_id: "smoke_user32", username: "user", is_active: true },
            { user_id: "smoke_user42", username: "user", is_active: true },
            { user_id: "smoke_user52", username: "user", is_active: true },
            { user_id: "smoke_user62", username: "user", is_active: true },
            { user_id: "smoke_user13", username: "user", is_active: true },
            { user_id: "smoke_user23", username: "user", is_active: true },
            { user_id: "smoke_user33", username: "user", is_active: true },
            { user_id: "smoke_user43", username: "user", is_active: true },
            { user_id: "smoke_user53", username: "user", is_active: true },
            { user_id: "smoke_user63", username: "user", is_active: true },
        ]
    }), {
        headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    });
    check(res, { "team_add 201": r => r.status === 201 });
    if (res.status != 201) console.log(res.status)

    // TEAM GET
    res = http.get(`${CONFIG.BASE_URL}/team/get?team_name=team_smoke${now}`, {
        headers: { "X-Api-Key": CONFIG.ADMIN_API_KEY }
    });
    check(res, { "team_get 200": r => r.status === 200 });
    if (res.status != 200) console.log(res.status)

    // PR CREATE
    res = http.post(`${CONFIG.BASE_URL}/pullRequest/create`, JSON.stringify({
        pull_request_id: `pr_smoke${now}`,
        pull_request_name: "pr_name",
        author_id: "smoke_user1"
    }), {
        headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    });
    check(res, { "pr_create 201": r => r.status === 201 });
    if (res.status != 201) console.log(res.status)

    // PR REASSIGN
    const reviewerId = res.json().pr.assigned_reviewers[0];

    const reassignRes = http.post(
        `${CONFIG.BASE_URL}/pullRequest/reassign`,
        JSON.stringify({
            pull_request_id: `pr_smoke${now}`,
            old_user_id: reviewerId
        }),
        {
            headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
        }
    );

    check(reassignRes, {
        "reassign 200": r => r.status === 200,
    });
    if (reassignRes.status != 200) console.log(reassignRes.status)

    // PR MERGE
    res = http.post(`${CONFIG.BASE_URL}/pullRequest/merge`, JSON.stringify({
        pull_request_id: `pr_smoke${now}`
    }), {
        headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    });
    check(res, { "pr_merge 200": r => r.status === 200 || r.status === 400 });
    if (res.status != 200 && res.status != 400) console.log(res.status)

    // USERS setIsActive
    res = http.post(`${CONFIG.BASE_URL}/users/setIsActive`, JSON.stringify({
        user_id: "smoke_user1",
        is_active: false
    }), {
        headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    });
    check(res, { "setIsActive 200": r => r.status === 200 });
    if (res.status != 200) console.log(res.status)

    // USERS getReview
    res = http.get(`${CONFIG.BASE_URL}/users/getReview?user_id=smoke_user1`, {
        headers: { "X-Api-Key": CONFIG.ADMIN_API_KEY }
    });
    check(res, { "getReview 200": r => r.status === 200 });
    if (res.status != 200) console.log(res.status)

    // USERS deactivateTeam
    res = http.post(`${CONFIG.BASE_URL}/team/deactivate`, JSON.stringify({
        team_name: `team_smoke${now}`
    }), {
        headers: { "Content-Type": "application/json", "X-Api-Key": CONFIG.ADMIN_API_KEY }
    });
    check(res, { "deactivateTeam 200": r => r.status === 200 });
    if (res.status != 200) console.log(res.status)

    sleep(1);
}