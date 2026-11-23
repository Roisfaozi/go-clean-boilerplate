import { check, fail, sleep } from "k6";
import http from "k6/http";

const base = (__ENV.BASE_URL || "http://127.0.0.1:8080/api/v1").replace(/\/$/, "");
export const options = {
	vus: 5,
	duration: "10s",
	thresholds: {
		http_req_failed: ["rate<0.01"],
		http_req_duration: ["p(95)<200"],
		checks: ["rate==1"],
	},
};

export function setup() {
	if (!__ENV.USERNAME || !__ENV.PASSWORD || !__ENV.ORG_ID) {
		fail("Set USERNAME, PASSWORD and ORG_ID for a test organization");
	}
	const login = http.post(`${base}/auth/login`,
		JSON.stringify({ username: __ENV.USERNAME, password: __ENV.PASSWORD }),
		{ headers: { "Content-Type": "application/json" } });
	if (login.status !== 200) fail(`Login failed (HTTP ${login.status})`);
	const token = login.json("data.access_token");
	if (!token) fail("Login returned no access token");
	return { Authorization: `Bearer ${token}`, "X-Organization-ID": __ENV.ORG_ID };
}

export default function (headers) {
	for (const path of ["/health", "/auth/me", "/organizations/me", "/projects"]) {
		const res = http.get(base + path, { headers });
		check(res, { [`GET ${path} returns 200`]: (r) => r.status === 200 });
	}
	if (__ITER === 0) {
		const key = crypto.randomUUID();
		const body = JSON.stringify({ name: `load-${key}`, domain: `${key}.example.test` });
		const params = { headers: { ...headers, "Content-Type": "application/json", "Idempotency-Key": key } };
		const ids = new Set();
		try {
			for (let attempt = 0; attempt < 2; attempt++) {
				const res = http.post(`${base}/projects`, body, params);
				if (!check(res, { "project created or replayed": (r) => r.status === 201 })) break;
				const id = res.json("data.id");
				if (!check(id, { "project ID returned": (value) => typeof value === "string" && !!value })) break;
				ids.add(id);
				if (attempt === 1) check(res, {
					"same project replayed": (r) => ids.size === 1 && r.headers["X-Cache-Lookup"] === "HIT - Idempotency Replay",
				});
			}
		} finally {
			for (const id of ids) {
				const res = http.del(`${base}/projects/${encodeURIComponent(id)}`, null, { headers, tags: { name: "DELETE /projects/:id" } });
				check(res, { "test project deleted": (r) => r.status === 200 });
			}
		}
	}
	sleep(1);
}

export function teardown(headers) {
	check(http.post(`${base}/auth/logout`, null, { headers }), {
		"test session revoked": (r) => r.status === 200,
	});
}
